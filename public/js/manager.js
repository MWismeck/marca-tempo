
  const employeeList = document.getElementById("employee-list");
  const editFields = document.getElementById("edit-fields");
  const modal = new bootstrap.Modal(document.getElementById("editModal"));
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
      const email = localStorage.getItem("employee_email");
      const res = await axios.get(`/employees?active=true`, {
        headers: { "X-User-Role": "manager" }
      });
      const employees = res.data["employees:"] || [];

      employeeList.innerHTML = employees.map(emp => `
        <tr>
          <td>${emp.name}</td>
          <td>${emp.email}</td>
          <td><button class="btn btn-sm btn-primary" onclick="editLogs('${emp.email}')">Editar</button></td>
        </tr>
      `).join("");
    } catch (err) {
      console.error(err);
      employeeList.innerHTML = "<tr><td colspan='3'>Erro ao carregar.</td></tr>";
    }
  }

  async function editLogs(email) {
    currentEmail = email;
    try {
      const res = await axios.get(`/time_logs?employee_email=${email}`);
      logsCache = res.data;

      if (!logsCache.length) return alert("Sem registros!");

      const latest = logsCache[0]; // Mais recente
      editFields.innerHTML = `
        ${formatInput("Entrada", latest.entry_time, "entry_time")}
        ${formatInput("Saída Almoço", latest.lunch_exit_time, "lunch_exit_time")}
        ${formatInput("Retorno Almoço", latest.lunch_return_time, "lunch_return_time")}
        ${formatInput("Saída", latest.exit_time, "exit_time")}
      `;
      modal.show();
    } catch (err) {
      console.error(err);
      alert("Erro ao buscar registros.");
    }
  }

  document.getElementById("edit-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const inputs = e.target.elements;
    const body = {
      entry_time: inputs.entry_time.value,
      lunch_exit_time: inputs.lunch_exit_time.value,
      lunch_return_time: inputs.lunch_return_time.value,
      exit_time: inputs.exit_time.value,
      editado_por_gerente: managerName,
      editado_em: new Date().toISOString()
    };

    try {
      const id = logsCache[0].id;
      await axios.put(`/time_logs/${id}`, body);
      alert("Alterações salvas!");
      modal.hide();
    } catch (err) {
      console.error(err);
      alert("Erro ao salvar.");
    }
  });

  document.getElementById("btn-export").addEventListener("click", () => {
    const email = localStorage.getItem("employee_email");
    window.open(`/time_logs/export?employee_email=${encodeURIComponent(email)}`, "_blank");
  });

  document.getElementById("btn-logout").addEventListener("click", () => {
    localStorage.clear();
    window.location.href = "index.html";
  });

  fetchEmployees();
