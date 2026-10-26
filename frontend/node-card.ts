/** Data needed to render one node in the Nodes View. */
interface NodeCardData {
  nodeId: string;
  state: Record<string, number> | string[] | number;
  disabled?: boolean;
}

function createNodeCard({ nodeId, state, disabled = false }: NodeCardData): HTMLElement {
  const card = document.createElement("article");
  card.className = disabled ? "node-card node-card--disabled" : "node-card";
  card.dataset.nodeId = nodeId;

  const isOpCounter = typeof state === "number";
  const isSet = Array.isArray(state);
  const stateEntries = isSet || isOpCounter ? [] : Object.entries(state);
  const total = isSet ? 0 : stateEntries.reduce((sum, [, value]) => sum + value, 0);
  const rows = isSet
    ? ""
    : stateEntries
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
  const hasRows = stateEntries.length > 0;
  const emptyText = "No replica state";
  const setBody = `
      <div class="node-card-caption">Elements</div>
      <div class="node-card-set">{${
        (state as string[]).length > 0
          ? ` ${(state as string[])
              .map((element) => `<span class="node-card-element" data-element="${escapeXml(element)}">${escapeXml(element)}</span>`)
              .join(", ")} `
          : " "
      }}</div>
    `;

  card.innerHTML = `
    <div class="node-card-header">
      <strong>Node ${nodeId}</strong>
      ${isSet || isOpCounter ? "" : `<span class="node-card-sum" title="Counter value (sum of state)">${total}</span>`}
      <div class="node-card-actions">
        <button class="node-card-action node-card-power${disabled ? " node-card-power--enable" : ""}" type="button" data-action="toggle-power" aria-label="${disabled ? "Enable" : "Disable"} node ${nodeId}" title="${disabled ? "Enable" : "Disable"} node"><svg class="node-card-power-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg></button>
      </div>
    </div>
    <div class="node-card-body">
      ${
        isOpCounter
          ? `<div class="node-card-caption">Value</div><div class="node-card-op-value" data-op-value>${state}</div>`
          : isSet
          ? setBody
          : `<table class="node-card-state-table">
        <tbody>
          ${hasRows ? rows : `<tr><td class="node-card-empty">${emptyText}</td></tr>`}
        </tbody>
      </table>`
      }
    </div>
  `;

  card
    .querySelector<HTMLButtonElement>('[data-action="toggle-power"]')
    ?.addEventListener("click", () => toggleNodePower(nodeId, disabled));

  return card;
}
