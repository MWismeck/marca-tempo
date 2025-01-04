document.getElementById('time-registration-component').innerHTML = `
    <div class="card p-4 mx-auto mt-5" style="max-width: 600px;">
        <h3 class="text-center color1">Bater Ponto</h3>
        <button id="punch-clock" class="btn color3 w-100">Registrar Ponto</button>
        <h4 class="text-center mt-4">Logs de Pontos</h4>
        <ul id="time-logs" class="list-group"></ul>
    </div>
`;

const fetchTimeLogs = async () => {
    try {
        const response = await axios.get('/time_logs/');
        const logs = response.data;
        const logsList = logs.map(log => `<li class="list-group-item">${log.timestamp}</li>`).join('');
        document.getElementById('time-logs').innerHTML = logsList;
    } catch (error) {
        alert('Erro ao buscar logs de pontos.');
    }
};

document.getElementById('punch-clock').addEventListener('click', async () => {
    try {
        const response = await axios.post('/time_logs/');
        alert('Ponto registrado com sucesso!');
        fetchTimeLogs();
    } catch (error) {
        alert('Erro ao registrar ponto.');
    }
});

fetchTimeLogs();
