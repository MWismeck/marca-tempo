document.getElementById('register-component').innerHTML = `
    <div class="card p-4 mx-auto mt-5" style="max-width: 500px;">
        <h3 class="text-center color1">Registrar Usuário</h3>
        <form id="register-form">
            <div class="mb-3">
                <label for="register-name" class="form-label">Nome</label>
                <input type="text" class="form-control" id="register-name" required>
            </div>
            <div class="mb-3">
                <label for="register-email" class="form-label">Email</label>
                <input type="email" class="form-control" id="register-email" required>
            </div>
            <div class="mb-3">
                <label for="register-password" class="form-label">Senha</label>
                <input type="password" class="form-control" id="register-password" required>
            </div>
            <button type="submit" class="btn color3 w-100">Registrar</button>
        </form>
    </div>
`;

document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    try {
        const response = await axios.post('/employee/', { name, email, password });
        alert('Usuário registrado com sucesso!');
    } catch (error) {
        alert('Erro ao registrar usuário.');
    }
});
