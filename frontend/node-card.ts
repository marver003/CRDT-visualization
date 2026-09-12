/** Data needed to render one node in the Nodes View. */
interface NodeCardData {
  nodeId: string;
  state: Record<string, number>;
}

/**
 * Creates a node card from data, so the same HTML structure is used for every
 * node. The card shows only the state vector, transposed to one row per
 * replica (rather than one column per replica) so it stays readable as more
 * nodes join. The row for the card's own replica is bolded, since that's the
 * one counter this node actually owns/increments - there's no separate "Own
 * counter" metric anymore.
 * For example: createNodeCard({ nodeId: "B", state: { A: 2, B: 3 } })
 */
function createNodeCard({ nodeId, state }: NodeCardData): HTMLElement {
  const card = document.createElement("article");
  card.className = "node-card";
  card.dataset.nodeId = nodeId;

  const stateEntries = Object.entries(state);
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
      <div class="node-card-actions">
        <button class="node-card-action" type="button" data-action="increment" aria-label="Increment node ${nodeId}" title="Increment node">+</button>
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
    .querySelector<HTMLButtonElement>('[data-action="increment"]')
    ?.addEventListener("click", () => incrementNode(nodeId));
  card
    .querySelector<HTMLButtonElement>('[data-action="remove"]')
    ?.addEventListener("click", () => removeNode(nodeId));

  return card;
}
