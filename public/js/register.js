const registerForm = document.getElementById('register-form');
registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const employee = {
        name: document.getElementById('register-name').value,
        cpf: document.getElementById('register-cpf').value,
        rg: document.getElementById('register-rg').value,
        email: document.getElementById('register-email').value,
        age: parseInt(document.getElementById('register-age').value, 10),
        workload: parseFloat(document.getElementById('register-workload').value),
        active: document.getElementById('register-active').value === "true"
    };

    const password = document.getElementById('register-password').value;
    try {
        // Cadastrar funcionário
        const employeeResponse = await axios.post('http://localhost:8080/employee/', employee);

        if (employeeResponse.status === 201) {
            // Cadastrar senha
            const passwordResponse = await axios.post('http://localhost:8080/login/password', {
                email: employee.email,
                password: password
            });

            if (passwordResponse.status === 200) {
                alert('Funcionário e senha cadastrados com sucesso!');
                registerForm.reset();
            }
        }
    } catch (err) {
        alert('Erro ao registrar funcionário ou senha.');
        console.error(err);
    }
});
