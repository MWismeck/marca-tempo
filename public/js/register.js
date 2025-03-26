 document.addEventListener("DOMContentLoaded", function () {
     const registerForm = document.getElementById('nregister');

     if (registerForm) {
         registerForm.addEventListener('submit', function (e) {
             e.preventDefault();
             window.location.href = "public/register.html";

         });
     }
 });
document.addEventListener("DOMContentLoaded", function () {
    console.log("register.js carregado!");

    const registerForm = document.getElementById('nregister');

    if (registerForm) {
        console.log("Formulário encontrado!");

        registerForm.addEventListener('submit', function (e) {
            e.preventDefault();
            console.log("Redirecionando para register.html...");
            window.location.href = "public/register.html";

        });
    } else {
        console.log("Formulário não encontrado!");
    }
});




const registerForm = document.getElementById('register-form');
// registerForm.addEventListener('submit', async (e) => {
//     e.preventDefault();

//     const employee = {
//         name: document.getElementById('register-name').value,
//         cpf: document.getElementById('register-cpf').value,
//         rg: document.getElementById('register-rg').value,
//         email: document.getElementById('register-email').value,
//         age: parseInt(document.getElementById('register-age').value, 10),
//         workload: parseFloat(document.getElementById('register-workload').value),
//         active: document.getElementById('register-active').value === "true"
//     };
//     console.log('Enviando funcionário:', employee);
//     const password = document.getElementById('register-password').value;
//     try {
//         // Create Employee
//         const employeeResponse = await axios.post('http://localhost:8080/employee/', employee);

//         if (employeeResponse.status === 200 || employeeResponse.status === 201) {
//             console.log('Employee created:', employeeResponse.data);

//             // Create or Update Password
//             const passwordResponse = await axios.post('http://localhost:8080/login/password', {
//                 email: employee.email,
//                 password: password
//             });

//             if (passwordResponse.status === 200) {
//                 console.log('Password updated:', passwordResponse.data);
//                 alert('Funcionário e senha cadastrados com sucesso!');
//                 registerForm.reset();
//             } else {
//                 console.error('Failed to update password:', passwordResponse);
//                 alert('Erro ao registrar a senha do funcionário.');
//             }
//         } else {
//             console.error('Failed to create employee:', employeeResponse);
//             alert('Erro ao registrar funcionário.');
//         }
//     } catch (err) {
//         console.error('Error during registration process:', err);
//         alert('Erro ao registrar funcionário ou senha. Verifique os logs para mais detalhes.');
//     }
// });
if (registerForm) {
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
            const employeeResponse = await axios.post('http://localhost:8080/employee/', employee);

            if (employeeResponse.status === 201) {
                const passwordResponse = await axios.post('http://localhost:8080/login/password', {
                    email: employee.email,
                    password: password
                });

                if (passwordResponse.status === 200) {
                    alert('Funcionário e senha cadastrados com sucesso! Redirecionando para o login...');
                    window.location.href = "index.html"; // Redireciona para o login
                }
            }
        } catch (error) {
            alert('Erro ao registrar o funcionário. Verifique os dados e tente novamente.');
            console.error(error);
        }
    });
}
});