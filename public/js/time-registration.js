// Botão de registro de ponto
document.getElementById('register-time-btn').addEventListener('click', async () => {
    const employeeId = localStorage.getItem('employee_id');

    if (!employeeId) {
        alert('Funcionário não autenticado. Faça login novamente.');
        return;
    }

    try {
        // Tenta buscar o último registro de ponto do funcionário
        const response = await axios.get('http://localhost:8080/time_logs', {
            params: { employee_id: employeeId },
        });

        const logs = response.data;

        if (logs && logs.length > 0) {
            const lastLog = logs[logs.length - 1]; // Último registro de ponto

            // Verifica se o último registro precisa ser atualizado
            if (!lastLog.exit_time) {
                // Atualiza o horário de saída
                const updateResponse = await axios.put(`http://localhost:8080/time_logs/${lastLog.id}`, {
                    exit_time: new Date().toISOString(), // Horário atual
                });

                if (updateResponse.status === 200) {
                    alert('Horário de saída registrado com sucesso!');
                    fetchTimeLogs(); // Atualiza a lista de registros
                }
                return;
            }
        }

        // Se não houver registros abertos, cria um novo
        const createResponse = await axios.post('http://localhost:8080/time_logs/', {
            employee_id: employeeId,
            entry_time: new Date().toISOString(), // Horário atual
        });

        if (createResponse.status === 201) {
            alert('Horário de entrada registrado com sucesso!');
            fetchTimeLogs(); // Atualiza a lista de registros
        }
    } catch (error) {
        alert('Erro ao registrar o ponto.');
        console.error(error);
    }
});

// Função para buscar e exibir registros de ponto
async function fetchTimeLogs() {
    const employeeId = localStorage.getItem('employee_id');

    if (!employeeId) {
        alert('Funcionário não autenticado. Faça login novamente.');
        return;
    }

    try {
        const response = await axios.get('http://localhost:8080/time_logs', {
            params: { employee_id: employeeId },
        });

        const logsContainer = document.getElementById('time-logs-container');
        logsContainer.innerHTML = response.data.map(log => `
            <div class="time-log border p-2 mb-2 rounded">
                <p><strong>Entrada:</strong> ${new Date(log.entry_time).toLocaleString()}</p>
                <p><strong>Saída:</strong> ${log.exit_time ? new Date(log.exit_time).toLocaleString() : '---'}</p>
            </div>
        `).join('');
    } catch (error) {
        alert('Erro ao buscar registros de ponto.');
        console.error(error);
    }
}

// Carrega os registros ao abrir a página
fetchTimeLogs();
