document.getElementById('login-component').innerHTML = `
    <div class="card p-4 mx-auto mt-5" style="max-width: 400px;">
        <h3 class="text-center color1">Login</h3>
        <form id="login-form">
            <div class="mb-3">
                <label for="login-email" class="form-label">Email</label>
                <input type="email" class="form-control" id="login-email" required>
            </div>
            <div class="mb-3">
                <label for="login-password" class="form-label">Senha</label>
                <input type="password" class="form-control" id="login-password" required>
            </div>
            <button type="submit" class="btn color3 w-100">Entrar</button>
        </form>
    </div>
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
});
