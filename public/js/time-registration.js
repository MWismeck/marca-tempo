document.addEventListener("DOMContentLoaded", () => {
  const email = localStorage.getItem("employee_email");
  const nome = localStorage.getItem("employee_name");
  const table = document.getElementById("time-table-body");

  if (!email || !nome || !table) {
    alert("Sessão expirada ou elementos ausentes.");
    window.location.href = "index.html";
    return;
  }

  // Mostrar nome do usuário
  const userNomeEl = document.getElementById("user-nome");
  if (userNomeEl) userNomeEl.innerText = nome;

  // Carrega pontos no HTML
  async function carregarPontos() {
    try {
      const res = await axios.get(`/time_logs/by_email/${email}`);
      table.innerHTML = "";

      res.data.forEach(log => {
        const row = document.createElement("tr");
        row.innerHTML = `
          <td>${log.date}</td>
          <td>${log.entry_time || "-"}</td>
          <td>${log.exit_time || "-"}</td>
        `;
        table.appendChild(row);
      });
    } catch (err) {
      console.error("Erro ao carregar pontos:", err);
      alert("Erro ao carregar pontos.");
    }
  }

  carregarPontos();

  // Botão para abrir modal
  const btnOpen = document.getElementById("btn-request-edit");
  if (btnOpen) {
    btnOpen.addEventListener("click", async () => {
      const res = await axios.get(`/time_logs/by_email/${email}`);
      const tbody = document.getElementById("request-table-body");
      tbody.innerHTML = "";

      res.data.forEach(log => {
        tbody.innerHTML += `
          <tr>
            <td><input type="radio" name="selected-log" value="${log.id}" data-date="${log.date}"></td>
            <td>${log.date}</td>
            <td>${log.entry_time || "-"}</td>
            <td>${log.exit_time || "-"}</td>
          </tr>
        `;
      });

      new bootstrap.Modal(document.getElementById("requestModal")).show();
    });
  }

  // Envia solicitação
  const form = document.getElementById("form-request-edit");
  if (form) {
    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      const motivo = document.getElementById("request-reason").value;
      const selected = document.querySelector("input[name='selected-log']:checked");

      if (!selected || motivo.trim().length < 3) {
        return alert("Selecione um ponto e escreva um motivo.");
      }

      const date = selected.getAttribute("data-date");

      try {
        await axios.post("/employee/request_change", {
          funcionario_email: email,
          data_solicitada: date,
          motivo
        });

        alert("Solicitação enviada com sucesso!");
        bootstrap.Modal.getInstance(document.getElementById("requestModal")).hide();
      } catch (err) {
        console.error("Erro ao enviar solicitação:", err);
        alert("Erro ao enviar solicitação.");
      }
    });
  }

  // Logout
  const logoutBtn = document.getElementById("btn-logout");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", () => {
      localStorage.clear();
      window.location.href = "index.html";
    });
  }
});
