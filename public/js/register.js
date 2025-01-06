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

    try {
        const response = await axios.post('http://localhost:8080/employee/', employee);

        if (response.status === 201) {
            alert('Funcionário registrado com sucesso!');
            registerForm.reset();
        }
    } catch (err) {
        alert('Erro ao registrar funcionário.');
        console.error(err);
    }
});
