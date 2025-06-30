document.addEventListener("DOMContentLoaded", () => {
  const employeeList = document.getElementById("employee-list");
  const editFields = document.getElementById("edit-fields");
  const modal = new bootstrap.Modal(document.getElementById("editModal"));
  const exportModal = new bootstrap.Modal(document.getElementById("exportModal"));
  const processModal = new bootstrap.Modal(document.getElementById("processRequestModal"));
  
  let logsCache = [];
  let currentEmail = "";
  let currentRequestId = null;
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

  // Carregar funcionários da empresa
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

  // Carregar solicitações
  async function loadRequests() {
    try {
      const managerEmail = localStorage.getItem("employee_email");
      if (!managerEmail) {
        alert("Sessão inválida. Faça login novamente.");
        window.location.href = "index.html";
        return;
      }

      console.log("Carregando solicitações para gerente:", managerEmail);
      const res = await axios.get(`http://localhost:8080/manager/requests?manager_email=${encodeURIComponent(managerEmail)}`);
      
      // Verificação de segurança para evitar erros
      const responseData = res.data || {};
      const pending = responseData.pending || [];
      const processed = responseData.processed || [];
      
      console.log("Solicitações carregadas:", { pending: pending.length, processed: processed.length });

      // Atualizar badges de notificação
      updateNotificationBadges(pending.length, processed.length);

      // Preencher tabela de pendentes
      fillPendingTable(pending);

      // Preencher tabela de histórico
      fillHistoryTable(processed);

    } catch (err) {
      console.error("Erro ao carregar solicitações:", err);
      document.getElementById("pending-requests-list").innerHTML = 
        "<tr><td colspan='5' class='text-center text-danger'>Erro ao carregar solicitações.</td></tr>";
      document.getElementById("history-requests-list").innerHTML = 
        "<tr><td colspan='6' class='text-center text-danger'>Erro ao carregar histórico.</td></tr>";
    }
  }

  // Atualizar badges de notificação
  function updateNotificationBadges(pendingCount, processedCount) {
    const pendingBadge = document.getElementById("pending-requests-badge");
    const pendingCountSpan = document.getElementById("pending-count");
    const pendingTabCount = document.getElementById("pending-tab-count");
    const historyTabCount = document.getElementById("history-tab-count");

    // Badge principal no cabeçalho
    if (pendingCount > 0) {
      pendingBadge.style.display = "inline";
      pendingCountSpan.textContent = pendingCount;
    } else {
      pendingBadge.style.display = "none";
    }

    // Badges nas abas
    pendingTabCount.textContent = pendingCount;
    historyTabCount.textContent = processedCount;
  }

  // Preencher tabela de solicitações pendentes
  function fillPendingTable(pending) {
    const tbody = document.getElementById("pending-requests-list");
    
    // Verificação de segurança
    if (!pending || !Array.isArray(pending)) {
      tbody.innerHTML = "<tr><td colspan='5' class='text-center'>Nenhuma solicitação pendente.</td></tr>";
      return;
    }
    
    if (pending.length === 0) {
      tbody.innerHTML = "<tr><td colspan='5' class='text-center'>Nenhuma solicitação pendente.</td></tr>";
      return;
    }

    tbody.innerHTML = pending.map(req => `
      <tr>
        <td>${req.funcionario_nome || req.funcionario_email}</td>
        <td>${new Date(req.data_solicitada).toLocaleDateString('pt-BR')}</td>
        <td class="text-truncate" style="max-width: 200px;" title="${req.motivo}">${req.motivo}</td>
        <td>${new Date(req.CreatedAt).toLocaleDateString('pt-BR')} ${new Date(req.CreatedAt).toLocaleTimeString('pt-BR')}</td>
        <td>
          <button class="btn btn-sm btn-warning" onclick="processRequest(${req.ID})">
            <i class="fas fa-cog"></i> Processar
          </button>
        </td>
      </tr>
    `).join("");
  }

  // Preencher tabela de histórico
  function fillHistoryTable(processed) {
    const tbody = document.getElementById("history-requests-list");
    
    // Verificação de segurança
    if (!processed || !Array.isArray(processed)) {
      tbody.innerHTML = "<tr><td colspan='6' class='text-center'>Nenhuma solicitação processada.</td></tr>";
      return;
    }
    
    if (processed.length === 0) {
      tbody.innerHTML = "<tr><td colspan='6' class='text-center'>Nenhuma solicitação processada.</td></tr>";
      return;
    }

    tbody.innerHTML = processed.map(req => {
      const statusClass = req.status === 'aprovado' ? 'text-success' : 'text-danger';
      const statusIcon = req.status === 'aprovado' ? 'fa-check' : 'fa-times';
      
      return `
        <tr>
          <td>${req.funcionario_nome || req.funcionario_email}</td>
          <td>${new Date(req.data_solicitada).toLocaleDateString('pt-BR')}</td>
          <td class="${statusClass}">
            <i class="fas ${statusIcon}"></i> ${req.status.charAt(0).toUpperCase() + req.status.slice(1)}
          </td>
          <td>${req.gerente_email || 'N/A'}</td>
          <td class="text-truncate" style="max-width: 200px;" title="${req.comentario_gerente || ''}">${req.comentario_gerente || 'Sem comentário'}</td>
          <td>${req.processado_em ? new Date(req.processado_em).toLocaleDateString('pt-BR') + ' ' + new Date(req.processado_em).toLocaleTimeString('pt-BR') : 'N/A'}</td>
        </tr>
      `;
    }).join("");
  }

  // Processar solicitação (aprovar/rejeitar)
  window.processRequest = async function(requestId) {
    try {
      currentRequestId = requestId;
      
      // Buscar detalhes da solicitação
      const managerEmail = localStorage.getItem("employee_email");
      const res = await axios.get(`http://localhost:8080/manager/requests?manager_email=${encodeURIComponent(managerEmail)}`);
      const { pending } = res.data;
      
      const request = pending.find(req => req.ID === requestId);
      if (!request) {
        alert("Solicitação não encontrada.");
        return;
      }

      // Preencher detalhes no modal
      document.getElementById("request-details").innerHTML = `
        <div class="card bg-light">
          <div class="card-body">
            <h6 class="card-title">Detalhes da Solicitação</h6>
            <p><strong>Funcionário:</strong> ${request.funcionario_nome || request.funcionario_email}</p>
            <p><strong>Data Solicitada:</strong> ${new Date(request.data_solicitada).toLocaleDateString('pt-BR')}</p>
            <p><strong>Motivo:</strong> ${request.motivo}</p>
            <p><strong>Solicitado em:</strong> ${new Date(request.CreatedAt).toLocaleDateString('pt-BR')} às ${new Date(request.CreatedAt).toLocaleTimeString('pt-BR')}</p>
          </div>
        </div>
      `;

      // Limpar comentário anterior
      document.querySelector('textarea[name="comentario"]').value = "";

      processModal.show();
    } catch (err) {
      console.error("Erro ao carregar detalhes da solicitação:", err);
      alert("Erro ao carregar detalhes da solicitação.");
    }
  };

  // Aprovar solicitação
  document.getElementById("approve-btn").addEventListener("click", async () => {
    await updateRequestStatus("aprovado");
  });

  // Rejeitar solicitação
  document.getElementById("reject-btn").addEventListener("click", async () => {
    await updateRequestStatus("rejeitado");
  });

  // Atualizar status da solicitação
  async function updateRequestStatus(status) {
    const comentario = document.querySelector('textarea[name="comentario"]').value.trim();
    
    if (!comentario || comentario.length < 5) {
      alert("O comentário é obrigatório e deve ter pelo menos 5 caracteres.");
      return;
    }

    try {
      const managerEmail = localStorage.getItem("employee_email");
      
      const body = {
        status: status,
        comentario_gerente: comentario,
        gerente_email: managerEmail
      };

      console.log("Processando solicitação:", { requestId: currentRequestId, body });

      await axios.put(`http://localhost:8080/manager/requests/${currentRequestId}/status`, body);
      
      alert(`Solicitação ${status} com sucesso!`);
      processModal.hide();
      
      // Recarregar solicitações
      await loadRequests();

      // Se aprovado, abrir modal de edição automaticamente
      if (status === "aprovado") {
        // Buscar detalhes da solicitação para obter o email do funcionário
        const managerEmail = localStorage.getItem("employee_email");
        const res = await axios.get(`http://localhost:8080/manager/requests?manager_email=${encodeURIComponent(managerEmail)}`);
        const allRequests = [...res.data.pending, ...res.data.processed];
        const request = allRequests.find(req => req.ID === currentRequestId);
        
        if (request) {
          setTimeout(() => {
            editLogs(request.funcionario_email);
          }, 500);
        }
      }

    } catch (err) {
      console.error("Erro ao processar solicitação:", err);
      const errorMsg = err.response?.data?.error || "Erro ao processar solicitação.";
      alert(errorMsg);
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

  // Atualização automática das solicitações a cada 30 segundos
  setInterval(loadRequests, 30000);

  // Carrega dados ao inicializar
  fetchEmployees();
  loadRequests();
});
