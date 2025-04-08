// Função para mostrar mensagem de status
function showStatusMessage(message, isError = false) {
    const statusElement = document.getElementById('status-message');
    statusElement.textContent = message;
    statusElement.className = `mt-2 alert ${isError ? 'alert-danger' : 'alert-success'}`;
    
    // Limpa a mensagem após 5 segundos
    setTimeout(() => {
        statusElement.textContent = '';
        statusElement.className = 'mt-2';
    }, 5000);
}

// Função para formatar data e hora
function formatDateTime(dateTimeStr) {
    // Check for null, undefined, or invalid date
    if (!dateTimeStr || new Date(dateTimeStr).toString() === 'Invalid Date') {
        return '---';
    }
    
    // Check for zero time (0001-01-01) which is Go's default time value
    if (dateTimeStr.includes('0001-01-01') || dateTimeStr.includes('01/01/0001')) {
        return '---';
    }
    
    // Check for very old dates that might be default values (before 2000)
    const date = new Date(dateTimeStr);
    if (date.getFullYear() < 2000) {
        return '---';
    }
    
    return date.toLocaleString('pt-BR');
}

// Função para formatar apenas a data
function formatDate(dateTimeStr) {
    // Check for null, undefined, or invalid date
    if (!dateTimeStr || new Date(dateTimeStr).toString() === 'Invalid Date') {
        return '---';
    }
    
    // Check for zero time (0001-01-01) which is Go's default time value
    if (dateTimeStr.includes('0001-01-01') || dateTimeStr.includes('01/01/0001')) {
        return '---';
    }
    
    // Check for very old dates that might be default values (before 2000)
    const date = new Date(dateTimeStr);
    if (date.getFullYear() < 2000) {
        return '---';
    }
    
    // Format the date as DD/MM/YYYY (day/month/year)
    const options = { day: '2-digit', month: '2-digit', year: 'numeric' };
    return date.toLocaleDateString('pt-BR', options);
}

// Função para formatar horas
function formatHours(hours) {
    if (hours === undefined || hours === null) {
        return '---';
    }
    return hours.toFixed(2) + 'h';
}

// Botão de registro de ponto
document.getElementById('register-time-btn').addEventListener('click', async () => {
    const employeeEmail = localStorage.getItem('employee_email');

    if (!employeeEmail) {
        showStatusMessage('Funcionário não autenticado. Faça login novamente.', true);
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return;
    }

    try {
        // Registra o ponto usando o endpoint punchTime
        const response = await axios.put('http://localhost:8080/time_logs/0', null, {
            params: { employee_email: employeeEmail }
        });

        if (response.status === 200 || response.status === 201) {
            const data = response.data;
            let message = '';
            
            // Determina qual registro foi feito com base nos dados retornados
            if (data.exit_time && !data.exit_time.includes('0001-01-01')) {
                message = 'Horário de saída registrado com sucesso!';
            } else if (data.lunch_return_time && !data.lunch_return_time.includes('0001-01-01')) {
                message = 'Retorno do almoço registrado com sucesso!';
            } else if (data.lunch_exit_time && !data.lunch_exit_time.includes('0001-01-01')) {
                message = 'Saída para almoço registrada com sucesso!';
            } else {
                message = 'Horário de entrada registrado com sucesso!';
            }
            
            showStatusMessage(message);
            fetchTimeLogs(); // Atualiza a tabela de registros
        }
    } catch (error) {
        console.error('Erro ao registrar ponto:', error);
        let errorMessage = 'Erro ao registrar o ponto.';
        
        if (error.response && error.response.data) {
            errorMessage = error.response.data;
        }
        
        showStatusMessage(errorMessage, true);
    }
});

// Função para buscar e exibir registros de ponto
async function fetchTimeLogs() {
    const employeeEmail = localStorage.getItem('employee_email');

    if (!employeeEmail) {
        showStatusMessage('Funcionário não autenticado. Faça login novamente.', true);
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return;
    }

    try {
        const response = await axios.get('http://localhost:8080/time_logs', {
            params: { employee_email: employeeEmail }
        });

        const tableBody = document.getElementById('time-logs-table-body');
        
        if (response.data && response.data.length > 0) {
            // Ordena os registros por data (mais recente primeiro)
            const sortedLogs = response.data.sort((a, b) => 
                new Date(b.log_date) - new Date(a.log_date)
            );
            
            tableBody.innerHTML = sortedLogs.map(log => `
                <tr>
                    <td>${formatDate(log.log_date)}</td>
                    <td>${formatDateTime(log.entry_time)}</td>
                    <td>${formatDateTime(log.lunch_exit_time)}</td>
                    <td>${formatDateTime(log.lunch_return_time)}</td>
                    <td>${formatDateTime(log.exit_time)}</td>
                    <td class="${log.balance >= 0 ? 'text-success' : 'text-danger'} fw-bold">${formatHours(log.balance)}</td>
                </tr>
            `).join('');
        } else {
            tableBody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center">Nenhum registro de ponto encontrado</td>
                </tr>
            `;
        }
    } catch (error) {
        console.error('Erro ao buscar registros:', error);
        let errorMessage = 'Erro ao buscar registros de ponto.';
        
        if (error.response && error.response.data) {
            errorMessage = error.response.data;
        }
        
        showStatusMessage(errorMessage, true);
        
        const tableBody = document.getElementById('time-logs-table-body');
        tableBody.innerHTML = `
            <tr>
                <td colspan="6" class="text-center text-danger">Erro ao carregar registros</td>
            </tr>
        `;
    }
}

// Verifica se o usuário está autenticado e exibe o nome do funcionário
function checkAuthentication() {
    const employeeEmail = localStorage.getItem('employee_email');
    const employeeName = localStorage.getItem('employee_name');
    
    if (!employeeEmail) {
        showStatusMessage('Você precisa fazer login para acessar esta página.', true);
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return false;
    }
    
    // Exibe o nome do funcionário se disponível
    if (employeeName) {
        const employeeNameElement = document.getElementById('employee-name');
        employeeNameElement.textContent = `Funcionário: ${employeeName}`;
        employeeNameElement.className = 'fs-5 fw-bold';
    }
    
    return true;
}

// Função para exportar registros para Excel
document.getElementById('export-excel-btn').addEventListener('click', async () => {
    const employeeEmail = localStorage.getItem('employee_email');

    if (!employeeEmail) {
        showStatusMessage('Funcionário não autenticado. Faça login novamente.', true);
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return;
    }

    try {
        // Cria um link temporário para download
        const link = document.createElement('a');
        link.href = `http://localhost:8080/time_logs/export?employee_email=${encodeURIComponent(employeeEmail)}`;
        link.setAttribute('download', 'registros_ponto.xlsx');
        document.body.appendChild(link);
        
        // Simula um clique no link para iniciar o download
        link.click();
        
        // Remove o link após o download
        document.body.removeChild(link);
        
        showStatusMessage('Download do arquivo Excel iniciado!');
    } catch (error) {
        console.error('Erro ao exportar para Excel:', error);
        showStatusMessage('Erro ao exportar registros para Excel.', true);
    }
});

// Botão de sair
document.getElementById('exit-btn').addEventListener('click', () => {
    // Limpa o localStorage e redireciona para a página de login
    localStorage.removeItem('employee_email');
    window.location.href = 'index.html';
});

// Botão de problemas
document.getElementById('problems-btn').addEventListener('click', () => {
    // Redireciona para o WhatsApp
    const phoneNumber = '5511999999999'; // Substitua pelo número correto
    const message = 'Olá, estou com problemas no sistema de registro de ponto.';
    const whatsappUrl = `https://wa.me/${phoneNumber}?text=${encodeURIComponent(message)}`;
    window.open(whatsappUrl, '_blank');
});

// Inicialização da página
document.addEventListener('DOMContentLoaded', () => {
    if (checkAuthentication()) {
        fetchTimeLogs();
    }
});
