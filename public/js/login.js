document.getElementById('login-form').addEventListener('submit', async (event) => {
    event.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await axios.post('http://localhost:8080/login', { email, password });

        if (response.status === 200) {
            alert('Login realizado com sucesso!');
            const employeeId = response.data.employee_id;

            // Armazena o ID no localStorage
            localStorage.setItem('employee_id', employeeId);

            // Redireciona para a página de ponto
            window.location.href = 'time-registration.html';
        }
    } catch (err) {
        alert('Erro ao realizar login. Verifique suas credenciais.');
        console.error(err);
    }
});
