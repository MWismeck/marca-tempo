document.addEventListener("DOMContentLoaded", () => {
  const employeeList = document.getElementById("employee-list");
  const editFields = document.getElementById("edit-fields");
  const modal = new bootstrap.Modal(document.getElementById("editModal"));
  const exportModal = new bootstrap.Modal(document.getElementById("exportModal"));
  let logsCache = [];
  let currentEmail = "";
  const managerName = localStorage.getItem("employee_name");

  function formatInput(label, value, name) {
    const val = value ? new Date(value).toISOString().slice(0, 16) : "";
    return `
      <div class="col-md-6">
        <label>${label}</label>
        <input type="datetime-local" class="form-control" name="${name}" value="${val}">
      </div>
    `;
  }

  async function fetchEmployees() {
    try {
      const managerEmail = localStorage.getItem("employee_email");
      if (!managerEmail) {
        alert("Sessão inválida. Faça login novamente.");
        window.location.href = "index.html";
        return;
      }

      console.log("Buscando funcionários da empresa do gerente:", managerEmail);
      const res = await axios.get(`http://localhost:8080/employees/?active=true&manager_email=${encodeURIComponent(managerEmail)}`);
      console.log("Resposta da API:", res.data);
      
      const employees = res.data["employees:"] || [];
      console.log("Funcionários encontrados:", employees);

      if (employees.length === 0) {
        employeeList.innerHTML = "<tr><td colspan='3' class='text-center'>Nenhum funcionário encontrado na sua empresa.</td></tr>";
        return;
      }

      employeeList.innerHTML = employees.map(emp => `
        <tr>
          <td>${emp.name}</td>
          <td>${emp.email}</td>
          <td><button class="btn btn-sm btn-primary" onclick="editLogs('${emp.email}')">Editar</button></td>
        </tr>
      `).join("");
      
      console.log("Tabela de funcionários atualizada com sucesso");
    } catch (err) {
      console.error("Erro ao carregar funcionários:", err);
      console.error("Detalhes do erro:", err.response?.data);
      if (err.response?.status === 401) {
        alert("Acesso negado. Verifique se você é um gerente válido.");
        window.location.href = "index.html";
      } else {
        employeeList.innerHTML = "<tr><td colspan='3' class='text-center text-danger'>Erro ao carregar funcionários da sua empresa.</td></tr>";
      }
    }
  }

  // Torna a função editLogs global para ser acessível pelo HTML
  window.editLogs = async function(email) {
    currentEmail = email;
    try {
      const res = await axios.get(`http://localhost:8080/time_logs?employee_email=${email}`);
      logsCache = res.data;

      if (!logsCache.length) return alert("Sem registros!");

      const latest = logsCache[0]; // Mais recente
      editFields.innerHTML = `
        ${formatInput("Entrada", latest.entry_time, "entry_time")}
        ${formatInput("Saída Almoço", latest.lunch_exit_time, "lunch_exit_time")}
        ${formatInput("Retorno Almoço", latest.lunch_return_time, "lunch_return_time")}
        ${formatInput("Saída", latest.exit_time, "exit_time")}
        <div class="col-12 mt-3">
          <label class="form-label"><strong>Motivo da Alteração *</strong></label>
          <textarea class="form-control" name="motivo_edicao" rows="3" required 
                    placeholder="Descreva o motivo da alteração (obrigatório)..."></textarea>
        </div>
      `;
      modal.show();
    } catch (err) {
      console.error(err);
      alert("Erro ao buscar registros.");
    }
  };

  // Formulário de edição
  document.getElementById("edit-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const inputs = e.target.elements;
    const managerEmail = localStorage.getItem("employee_email");
    
    // Validação do motivo
    const motivo = inputs.motivo_edicao.value.trim();
    if (!motivo || motivo.length < 5) {
      alert("O motivo da alteração é obrigatório e deve ter pelo menos 5 caracteres.");
      return;
    }

    const body = {
      entry_time: inputs.entry_time.value,
      lunch_exit_time: inputs.lunch_exit_time.value,
      lunch_return_time: inputs.lunch_return_time.value,
      exit_time: inputs.exit_time.value,
      motivo_edicao: motivo,
      manager_email: managerEmail
    };

    try {
      const id = logsCache[0].ID || logsCache[0].id;
      console.log("Enviando edição:", body);
      
      await axios.put(`http://localhost:8080/time_logs/${id}/manual_edit`, body);
      alert("Alterações salvas com sucesso!");
      modal.hide();
      
      // Limpar o formulário
      inputs.motivo_edicao.value = "";
    } catch (err) {
      console.error("Erro ao salvar:", err);
      const errorMsg = err.response?.data || "Erro ao salvar alterações.";
      alert(errorMsg);
    }
  });

  // Botão exportar por período
  const btnExportRange = document.getElementById("btn-export-range");
  if (btnExportRange) {
    btnExportRange.addEventListener("click", () => {
      exportModal.show();
    });
  }

  // Formulário de exportação por período
  const formExport = document.getElementById("form-export");
  if (formExport) {
    formExport.addEventListener("submit", (e) => {
      e.preventDefault();
      const email = document.getElementById("export-email").value;
      const start = document.getElementById("export-start").value;
      const end = document.getElementById("export-end").value;

      if (!email || !start || !end) {
        alert("Preencha todos os campos!");
        return;
      }

      const url = `http://localhost:8080/time_logs/export_range?employee_email=${encodeURIComponent(email)}&start=${start}&end=${end}`;
      window.open(url, "_blank");
      exportModal.hide();
    });
  }

  // Logout
  const logoutBtn = document.getElementById("btn-logout");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", () => {
      if (confirm("Deseja realmente sair?")) {
        localStorage.clear();
        window.location.href = "index.html";
      }
    });
  }

  // Carrega funcionários ao inicializar
  fetchEmployees();
});
