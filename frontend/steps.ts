const nextStepButton = document.querySelector(".btn-step") as HTMLButtonElement;
const queueContent = document.querySelector(".queue-content") as HTMLDivElement;

type Step = {
  stepId: number;
  type: string;
  replicaId?: number;
  from?: number;
  to?: number;
};

async function renderStepsQueue() {
  const response: { steps?: Step[] } = await getStepsAPI();

  queueContent.innerHTML = "";

  if (!response.steps || response.steps.length === 0) {
    queueContent.innerHTML = `
            <div class="queue-empty">
                <div class="queue-icon">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>

                <h3>No steps in queue</h3>
                <p>Perform some operations<br>to see steps here.</p>
            </div>
        `;

    return;
  }

  const list = document.createElement("div");
  list.className = "steps-list";

  response.steps.forEach((step) => {
    const item = document.createElement("div");
    item.className = "step-item";

    let description;

    switch (step.type) {
      case "create_node":
        description = `Create node ${step.replicaId}`;
        break;

      case "increment":
        description = `Increment node ${step.replicaId}`;
        break;

      case "send_gossip":
        description = `Send gossip ${step.from} → ${step.to}`;
        break;

      case "remove_node":
        description = `Remove node ${step.replicaId}`;
        break;

      default:
        description = step.type;
    }

    item.innerHTML = `
            <span class="step-id">${step.stepId}</span>
            <span class="step-description">${description}</span>
        `;

    list.appendChild(item);
  });

  queueContent.appendChild(list);
}

async function executeNextStep() {
  await executeNextStepAPI();
  await renderState();
}

nextStepButton.addEventListener("click", executeNextStep);
