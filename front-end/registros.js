// Exemplo de dados simulados
const records = [
    { day: "Seg - 23/12", entryExit: "07:38", extra: "00:00", missing: "00:00", balance: "00:00", request: "" },
    { day: "Dom - 22/12", entryExit: "Nenhum ponto", extra: "00:00", missing: "00:00", balance: "-02:12", request: "" },
    { day: "Sáb - 21/12", entryExit: "Nenhum ponto", extra: "00:00", missing: "00:00", balance: "-02:12", request: "" },
  ];
  
  // Popula a tabela ao carregar a página
  function populateTable() {
    const tableBody = document.getElementById("records");
    records.forEach((record) => {
      const row = document.createElement("tr");
      row.innerHTML = `
        <td><span>&#10004;</span></td>
        <td>${record.day}</td>
        <td>${record.entryExit}</td>
        <td>${record.extra}</td>
        <td>${record.missing}</td>
        <td>${record.balance}</td>
        <td>${record.request}</td>
      `;
      tableBody.appendChild(row);
    });
  }
  
  document.addEventListener("DOMContentLoaded", () => {
    populateTable();
  
    // Lógica para o botão "Bater Ponto"
    document.getElementById("baterPontoButton").addEventListener("click", () => {
      alert("Ponto registrado com sucesso!");
    });
  });
  