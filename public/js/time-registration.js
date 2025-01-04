// Renderizar o componente de registro de ponto
document.getElementById('time-registration-component').innerHTML = `
    <div class="card p-4 mx-auto mt-5" style="max-width: 700px;">
        <h3 class="text-center text-primary">Registro de Ponto</h3>
        <form id="time-registration-form">
            <div class="mb-3">
                <label for="employee-id" class="form-label">ID do Funcionário</label>
                <input type="number" class="form-control" id="employee-id" placeholder="Ex.: 123" required>
            </div>
            <div class="mb-3">
                <label for="entry-time" class="form-label">Horário de Entrada</label>
                <input type="datetime-local" class="form-control" id="entry-time" required>
            </div>
            <div class="mb-3">
                <label for="lunch-exit-time" class="form-label">Saída para Almoço</label>
                <input type="datetime-local" class="form-control" id="lunch-exit-time">
            </div>
            <div class="mb-3">
                <label for="lunch-return-time" class="form-label">Retorno do Almoço</label>
                <input type="datetime-local" class="form-control" id="lunch-return-time">
            </div>
            <div class="mb-3">
                <label for="exit-time" class="form-label">Horário de Saída</label>
                <input type="datetime-local" class="form-control" id="exit-time">
            </div>
            <div class="mb-3">
                <label for="workload" class="form-label">Carga Horária (em horas)</label>
                <input type="number" step="0.1" class="form-control" id="workload" placeholder="Ex.: 8">
            </div>
            <button type="submit" class="btn btn-primary w-100">Registrar Ponto</button>
        </form>
    </div>
`;

// Adicionar funcionalidade ao formulário
const form = document.getElementById('time-registration-form');
form.addEventListener('submit', async (event) => {
    event.preventDefault();

    const timeLog = {
        employee_id: parseInt(document.getElementById('employee-id').value, 10),
        entry_time: document.getElementById('entry-time').value,
        lunch_exit_time: document.getElementById('lunch-exit-time').value || null,
        lunch_return_time: document.getElementById('lunch-return-time').value || null,
        exit_time: document.getElementById('exit-time').value || null,
        workload: parseFloat(document.getElementById('workload').value) || 0
    };

    try {
        const response = await axios.post('/time-log/', timeLog);

        if (response.status === 201) {
            alert('Registro de ponto realizado com sucesso!');
            form.reset();
        }
    } catch (error) {
        alert('Erro ao registrar o ponto. Verifique os dados e tente novamente.');
        console.error(error);
    }
});
