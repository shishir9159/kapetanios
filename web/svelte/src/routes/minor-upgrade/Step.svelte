<script lang="ts">
    function toggleStep(stepNumber) {
        const content = document.getElementById(`step-${stepNumber}`);

        // Hide all step contents
        document.querySelectorAll('.step-content').forEach(step => {
            if (step.id !== `step-${stepNumber}`) {
                // step.style.display = 'none';
            }
        });

        // Toggle the clicked step
        content.style.display = content.style.display === "block" ? "none" : "block";
    }
</script>

<div class="step-container">
    <div class="step" onclick={toggleStep(1)}>
        <div class="step-number">1</div>
        <div class="step-title">Step 1</div>
    </div>
    <div class="step-content" id="step-1">
        <p>This is the description for Step 1.</p>
    </div>

    <div class="step" onclick={toggleStep(2)}>
        <div class="step-number">2</div>
        <div class="step-title">Step 2</div>
    </div>
    <div class="step-content" id="step-2">
        <p>This is the description for Step 2.</p>
    </div>

    <div class="step" onclick={toggleStep(3)}>
        <div class="step-number">3</div>
        <div class="step-title">Step 3</div>
    </div>
    <div class="step-content" id="step-3">
        <p>This is the description for Step 3.</p>
    </div>

    <div class="step" onclick={toggleStep(4)}>
        <div class="step-number">4</div>
        <div class="step-title">Step 4</div>
    </div>
    <div class="step-content" id="step-4">
        <p>This is the description for Step 4.</p>
    </div>

    <div class="step" onclick={toggleStep(5)}>
        <div class="step-number">5</div>
        <div class="step-title">Step 5</div>
    </div>
    <div class="step-content" id="step-5">
        <p>This is the description for Step 5.</p>
    </div>
</div>

<style>
    .step-container {
        font-family: Arial, sans-serif;
        background-color: #f4f4f4;
        /*display: flex;*/
        justify-content: center;
        align-items: center;
        height: 100vh;
        margin: 0;
        width: 300px;
        padding: 20px;
        border-radius: 10px;
        box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
    }
    .step {
        display: flex;
        align-items: center;
        cursor: pointer;
        padding: 10px;
        border-bottom: 1px solid #ddd;
        transition: background 0.3s;
    }
    .step:last-child {
        border-bottom: none;
    }
    .step:hover {
        background: #f0f0f0;
    }
    .step-number {
        width: 30px;
        height: 30px;
        border-radius: 50%;
        background: #007bff;
        color: white;
        display: flex;
        justify-content: center;
        align-items: center;
        font-weight: bold;
        margin-right: 15px;
    }
    .step-title {
        font-size: 16px;
        font-weight: bold;
    }
    .step-content {
        display: none;
        padding: 10px;
        background: #e9f5ff;
        border-left: 4px solid #007bff;
        margin: 5px 0;
        border-radius: 5px;
    }
</style>