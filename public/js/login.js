document.getElementById('login-form').addEventListener('submit', async (event) => {
    event.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await axios.post('http://localhost:8080/login', { email, password });

        if (response.status === 200) {
            // Armazena o email no localStorage (usado para identificar o funcionário)
            localStorage.setItem('employee_email', email);
            
            // Armazena também o ID se estiver disponível na resposta
            if (response.data && response.data.employee_id) {
                localStorage.setItem('employee_id', response.data.employee_id);
            }
            
            // Armazena o nome do funcionário se disponível
            if (response.data && response.data.employee_name) {
                localStorage.setItem('employee_name', response.data.employee_name);
            }

            // Mostra mensagem de sucesso
            const messageElement = document.createElement('div');
            messageElement.className = 'alert alert-success mt-3';
            messageElement.textContent = 'Login realizado com sucesso! Redirecionando...';
            document.getElementById('login-form').appendChild(messageElement);

            // Redireciona para a página de ponto após um breve delay
            setTimeout(() => {
                window.location.href = 'time-registration.html';
            }, 1500);
        }
    } catch (err) {
        // Mostra mensagem de erro
        const messageElement = document.createElement('div');
        messageElement.className = 'alert alert-danger mt-3';
        
        if (err.response && err.response.data) {
            messageElement.textContent = err.response.data;
        } else {
            messageElement.textContent = 'Erro ao realizar login. Verifique suas credenciais.';
        }
        
        // Remove qualquer mensagem anterior
        const previousMessage = document.querySelector('#login-form .alert');
        if (previousMessage) {
            previousMessage.remove();
        }
        
        document.getElementById('login-form').appendChild(messageElement);
        console.error(err);
    }
});
