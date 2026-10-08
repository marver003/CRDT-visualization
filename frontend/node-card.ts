/** Data needed to render one node in the Nodes View. */
interface NodeCardData {
  nodeId: string;
  state: Record<string, number>;
}

function createNodeCard({ nodeId, state }: NodeCardData): HTMLElement {
  const card = document.createElement("article");
  card.className = "node-card";
  card.dataset.nodeId = nodeId;

  const stateEntries = Object.entries(state);
  const total = stateEntries.reduce((sum, [, value]) => sum + value, 0);
  const rows = stateEntries
    .map(([replicaId, value]) => {
      const isOwn = replicaId === nodeId;
      return `
        <tr class="${isOwn ? "node-card-state-own" : ""}">
          <th scope="row">${replicaId}</th>
          <td data-replica-id="${replicaId}">${value}</td>
        </tr>
      `;
    })
    .join("");

  card.innerHTML = `
    <div class="node-card-header">
      <strong>Node ${nodeId}</strong>
      <span class="node-card-sum" title="Counter value (sum of state)">${total}</span>
      <div class="node-card-actions">
        <button class="node-card-action node-card-remove" type="button" data-action="remove" aria-label="Remove node ${nodeId}" title="Remove node">&#128465;</button>
      </div>
    </div>
    <div class="node-card-body">
      <table class="node-card-state-table">
        <tbody>
          ${stateEntries.length > 0 ? rows : '<tr><td class="node-card-empty">No replica state</td></tr>'}
        </tbody>
      </table>
    </div>
  `;

  card
    .querySelector<HTMLButtonElement>('[data-action="remove"]')
    ?.addEventListener("click", () => removeNode(nodeId));

  return card;
}
