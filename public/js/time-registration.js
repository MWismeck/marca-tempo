document.addEventListener("DOMContentLoaded", () => {
    const employeeEmail = localStorage.getItem("employee_email");
    const employeeName = localStorage.getItem("employee_name");
    const role = localStorage.getItem("role");
    const activeSession = localStorage.getItem("session_active");

    if (!employeeEmail || !employeeName || !role || activeSession !== "true") {
        alert("Sessão inválida ou expirada. Faça login novamente.");
        setTimeout(() => {
            window.location.href = "index.html";
        }, 1500);
        return;
    }

    const email = employeeEmail;
    const nome = employeeName;
    const table = document.getElementById("time-logs-table-body");

    // Mostrar nome do usuário
    const userNomeEl = document.getElementById("employee-name");
    if (userNomeEl) userNomeEl.innerText = `Bem-vindo, ${nome}`;

    // Carrega pontos no HTML
    async function carregarPontos() {
        try {
            const res = await axios.get(`http://localhost:8080/time_logs?employee_email=${encodeURIComponent(email)}`);
            table.innerHTML = "";

            if (!res.data || res.data.length === 0) {
                table.innerHTML = `<tr><td colspan="6" class="text-center">Nenhum ponto registrado ainda.</td></tr>`;
                return;
            }

            res.data.forEach(log => {
                const row = document.createElement("tr");
                const logDate = new Date(log.log_date).toLocaleDateString('pt-BR');
                const entryTime = log.entry_time ? new Date(log.entry_time).toLocaleTimeString('pt-BR', {hour: '2-digit', minute: '2-digit'}) : "-";
                const lunchExitTime = log.lunch_exit_time ? new Date(log.lunch_exit_time).toLocaleTimeString('pt-BR', {hour: '2-digit', minute: '2-digit'}) : "-";
                const lunchReturnTime = log.lunch_return_time ? new Date(log.lunch_return_time).toLocaleTimeString('pt-BR', {hour: '2-digit', minute: '2-digit'}) : "-";
                const exitTime = log.exit_time ? new Date(log.exit_time).toLocaleTimeString('pt-BR', {hour: '2-digit', minute: '2-digit'}) : "-";
                const balance = log.balance ? log.balance.toFixed(2) : "0.00";
                
                row.innerHTML = `
                    <td>${logDate}</td>
                    <td>${entryTime}</td>
                    <td>${lunchExitTime}</td>
                    <td>${lunchReturnTime}</td>
                    <td>${exitTime}</td>
                    <td class="${log.balance >= 0 ? 'text-success' : 'text-danger'}">${balance}h</td>
                `;
                table.appendChild(row);
            });
        } catch (err) {
            console.error("Erro ao carregar pontos:", err);
            table.innerHTML = `<tr><td colspan="6" class="text-center text-danger">Erro ao carregar pontos.</td></tr>`;
        }
    }

    // Registrar ponto
    const registerBtn = document.getElementById("register-time-btn");
    if (registerBtn) {
        registerBtn.addEventListener("click", async () => {
            try {
                const statusDiv = document.getElementById("status-message");
                statusDiv.innerHTML = '<div class="alert alert-info">Registrando ponto...</div>';
                
                const res = await axios.put(`http://localhost:8080/time_logs/1?employee_email=${encodeURIComponent(email)}`);
                
                if (res.status === 200 || res.status === 201) {
                    statusDiv.innerHTML = '<div class="alert alert-success">Ponto registrado com sucesso!</div>';
                    setTimeout(() => {
                        statusDiv.innerHTML = '';
                    }, 3000);
                    carregarPontos(); // Recarrega a tabela
                }
            } catch (err) {
                console.error("Erro ao registrar ponto:", err);
                const statusDiv = document.getElementById("status-message");
                statusDiv.innerHTML = '<div class="alert alert-danger">Erro ao registrar ponto. Tente novamente.</div>';
                setTimeout(() => {
                    statusDiv.innerHTML = '';
                }, 5000);
            }
        });
    }

    // Exportar Excel
    const exportBtn = document.getElementById("export-excel-btn");
    if (exportBtn) {
        exportBtn.addEventListener("click", () => {
            window.open(`http://localhost:8080/time_logs/export?employee_email=${encodeURIComponent(email)}`, "_blank");
        });
    }

    // Botão de suporte (WhatsApp)
    const problemsBtn = document.getElementById("problems-btn");
    if (problemsBtn) {
        problemsBtn.addEventListener("click", () => {
            const message = encodeURIComponent("Olá! Preciso de ajuda com o sistema de ponto.");
            window.open(`https://wa.me/5511999999999?text=${message}`, "_blank");
        });
    }

    // Botão sair
    const exitBtn = document.getElementById("exit-btn");
    if (exitBtn) {
        exitBtn.addEventListener("click", () => {
            if (confirm("Deseja realmente sair do sistema?")) {
                localStorage.clear();
                window.location.href = "index.html";
            }
        });
    }

    // Botão para abrir modal de solicitação
    const btnRequestEdit = document.getElementById("btn-request-edit");
    if (btnRequestEdit) {
        btnRequestEdit.addEventListener("click", async () => {
            try {
                const res = await axios.get(`http://localhost:8080/time_logs?employee_email=${encodeURIComponent(email)}`);
                
                if (!res.data || res.data.length === 0) {
                    alert("Você ainda não possui pontos registrados para solicitar alteração.");
                    return;
                }

                // Preenche a data atual no campo
                const today = new Date().toISOString().split('T')[0];
                document.getElementById("request-date").value = today;
                
                // Mostra o modal
                const modal = new bootstrap.Modal(document.getElementById("requestModal"));
                modal.show();
            } catch (err) {
                alert("Erro ao carregar pontos para solicitação.");
                console.error(err);
            }
        });
    }

    // Envia solicitação de alteração
    const formRequestEdit = document.getElementById("form-request-edit");
    if (formRequestEdit) {
        formRequestEdit.addEventListener("submit", async (e) => {
            e.preventDefault();
            const motivo = document.getElementById("request-reason").value;
            const dataSolicitada = document.getElementById("request-date").value;

            if (!dataSolicitada || motivo.trim().length < 3) {
                return alert("Preencha a data e escreva um motivo com pelo menos 3 caracteres.");
            }

            try {
                await axios.post("http://localhost:8080/employee/request_change", {
                    funcionario_email: email,
                    data_solicitada: new Date(dataSolicitada).toISOString(),
                    motivo: motivo
                });

                alert("Solicitação enviada com sucesso!");
                const modal = bootstrap.Modal.getInstance(document.getElementById("requestModal"));
                modal.hide();
                
                // Limpa o formulário
                document.getElementById("request-reason").value = "";
                document.getElementById("request-date").value = "";
            } catch (err) {
                console.error("Erro ao enviar solicitação:", err);
                alert("Erro ao enviar solicitação.");
            }
        });
    }

    // Carrega pontos ao inicializar
    carregarPontos();
});
