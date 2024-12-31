


document.getElementById("loginButton").addEventListener("click", () => {
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
  
    // Simples validação de login (pode ser conectado a um backend)
    if (username === "admin" && password === "1234") {
      // Redireciona para a página de registros
      window.location.href = "registros.html";
    } else {
      alert("Usuário ou senha incorretos!");
    }
  });
  