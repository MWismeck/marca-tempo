
`;

document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await axios.post('/api/login', { email, password });
        alert('Login bem-sucedido!');
        // Redirecionar para a página principal
    } catch (error) {
        alert('Erro no login. Verifique suas credenciais.');
    }
}); `
