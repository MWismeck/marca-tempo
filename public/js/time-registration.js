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
    if (!dateTimeStr || new Date(dateTimeStr).toString() === 'Invalid Date') {
        return '---';
    }
    return new Date(dateTimeStr).toLocaleString('pt-BR');
}

// Função para formatar apenas a data
function formatDate(dateTimeStr) {
    if (!dateTimeStr || new Date(dateTimeStr).toString() === 'Invalid Date') {
        return '---';
    }
    return new Date(dateTimeStr).toLocaleDateString('pt-BR');
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
                    <td>${formatHours(log.extra_hours)}</td>
                    <td>${formatHours(log.missing_hours)}</td>
                    <td class="${log.balance >= 0 ? 'text-success' : 'text-danger'}">${formatHours(log.balance)}</td>
                </tr>
            `).join('');
        } else {
            tableBody.innerHTML = `
                <tr>
                    <td colspan="8" class="text-center">Nenhum registro de ponto encontrado</td>
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
                <td colspan="8" class="text-center text-danger">Erro ao carregar registros</td>
            </tr>
        `;
    }
}

// Verifica se o usuário está autenticado
function checkAuthentication() {
    const employeeEmail = localStorage.getItem('employee_email');
    
    if (!employeeEmail) {
        showStatusMessage('Você precisa fazer login para acessar esta página.', true);
        setTimeout(() => {
            window.location.href = 'index.html';
        }, 2000);
        return false;
    }
    
    return true;
}

// Inicialização da página
document.addEventListener('DOMContentLoaded', () => {
    if (checkAuthentication()) {
        fetchTimeLogs();
    }
});
