document.getElementById('login-form').addEventListener('submit', async (event) => {
    event.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await axios.post('http://localhost:8080/login', { email, password });

        if (response.status === 200) {
    localStorage.setItem('employee_email', response.data.employee_email);
    localStorage.setItem('employee_id', response.data.employee_id || "");
    localStorage.setItem('employee_name', response.data.employee_name || "");
    localStorage.setItem('role', response.data.role || "");

    const messageElement = document.createElement('div');
    messageElement.className = 'alert alert-success mt-3';
    messageElement.textContent = 'Login realizado com sucesso! Redirecionando...';
    document.getElementById('login-form').appendChild(messageElement);

    if (response.data.role === "manager") {
        const escolha = confirm("Você deseja acessar o painel do gerente?\nClique em 'Cancelar' para registrar ponto como funcionário.");
        window.location.href = escolha ? "manager.html" : "time-registration.html";
    } else if (response.data.role === "admin") {
        window.location.href = "admin.html";
    } else {
        window.location.href = "time-registration.html";
    }
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
