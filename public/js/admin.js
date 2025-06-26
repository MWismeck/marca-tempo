

  // InputMask
  Inputmask({ mask: "999.999.999-99" }).mask("#manager-cpf");
  Inputmask({ mask: "99.999.999-9" }).mask("#manager-rg");
  Inputmask({ mask: "99.999.999/9999-99" }).mask("#company-cnpj");
  Inputmask({ mask: "(99) 9999-9999[9]" }).mask("#company-phone");

  // Cadastro de empresa
  document.getElementById("form-company").addEventListener("submit", async (e) => {
    e.preventDefault();
    const body = {
      name: document.getElementById("company-name").value,
      cnpj: document.getElementById("company-cnpj").value,
      email: document.getElementById("company-email").value,
      fone: document.getElementById("company-phone").value,
      active: document.getElementById("company-active").value === "true"
    };

    try {
      await axios.post("/admin/create_company", body, {
        headers: { "X-User-Role": "admin" }
      });
      alert("Empresa cadastrada com sucesso!");
    } catch (err) {
      alert("Erro ao cadastrar empresa.");
      console.error(err);
    }
  });

  // Cadastro de gerente
  document.getElementById("form-manager").addEventListener("submit", async (e) => {
    e.preventDefault();
    const body = {
      name: document.getElementById("manager-name").value,
      email: document.getElementById("manager-email").value,
      cpf: document.getElementById("manager-cpf").value,
      rg: document.getElementById("manager-rg").value,
      age: parseInt(document.getElementById("manager-age").value),
      workload: parseFloat(document.getElementById("manager-workload").value),
      active: document.getElementById("manager-active").value === "true",
      password: document.getElementById("manager-password").value,
      company_cnpj: document.getElementById("manager-cnpj").value
    };

    try {
      await axios.post("/admin/create_manager", body, {
        headers: { "X-User-Role": "admin" }
      });

      await axios.post("/login/password", {
        email: body.email,
        password: body.password
      });

      alert("Gerente cadastrado com sucesso!");
    } catch (err) {
      alert("Erro ao cadastrar gerente.");
      console.error(err);
    }
  });

  // Listagem
  document.getElementById("btn-load-data").addEventListener("click", async () => {
    try {
      const [companies, managers] = await Promise.all([
        axios.get("/admin/companies", { headers: { "X-User-Role": "admin" } }),
        axios.get("/admin/managers", { headers: { "X-User-Role": "admin" } })
      ]);

      const div = document.getElementById("list-data");
      div.innerHTML = `
        <h5>Empresas:</h5>
        <ul>${companies.data.map(c => `<li><strong>${c.name}</strong> - CNPJ: ${c.cnpj}</li>`).join("")}</ul>
        <h5>Gerentes:</h5>
        <ul>${managers.data.map(m => `<li>${m.name} (${m.email}) - CNPJ: ${m.company_cnpj}</li>`).join("")}</ul>
      `;
    } catch (err) {
      alert("Erro ao carregar dados.");
      console.error(err);
    }
  });
