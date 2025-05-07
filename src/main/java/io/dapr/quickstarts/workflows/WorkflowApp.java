package io.dapr.quickstarts.workflows;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RestController;
import io.dapr.workflows.runtime.WorkflowRuntime;
import io.dapr.workflows.runtime.WorkflowRuntimeStatus;
import io.dapr.workflows.client.DaprWorkflowClient;
import io.dapr.workflows.client.WorkflowInstanceStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import io.dapr.quickstarts.workflows.models.*;

@SpringBootApplication
@RestController
public class WorkflowApp {

  @Autowired
  private WorkflowRuntime workflowRuntime;

  private static final Logger logger = LoggerFactory.getLogger(WorkflowApp.class);
  private DaprWorkflowClient workflowClient = new DaprWorkflowClient();

  public static void main(String[] args) {
    SpringApplication.run(WorkflowApp.class, args);
  }

  @PostMapping("/workflow/start")
  public ResponseEntity<String> startWorkflow(@RequestBody OrderPayload order) {
    logger.info("Received request to start workflow for item: {} with quantity: {}", order.getItemName(),
        order.getQuantity());

    if (workflowRuntime == null) {
      logger.error("Workflow runtime is not initialized");
      throw new IllegalStateException("Workflow runtime is not initialized");
    }

    // Run the workflow
    try {
      String instanceId = workflowClient.scheduleNewWorkflow(OrderProcessingWorkflow.class, order);
      logger.info("Workflow execution started successfully for item: {} {}", order.getQuantity(), order.getItemName());
      return ResponseEntity.ok(instanceId);
    } catch (Exception e) {
      logger.error("Error starting workflow for item: {} {}", order.getQuantity(), order.getItemName(), e);
      return ResponseEntity.internalServerError().body("Failed to start workflow: " + e.getMessage());
    }
  }

  @GetMapping("/workflow/status/{workflowId}")
  public ResponseEntity<String> getWorkflowStatus(@PathVariable String workflowId) {
    try {
      WorkflowInstanceStatus status = workflowClient.getInstanceState(workflowId, true);
      if (status == null) {
        logger.error("Workflow not found for ID: {}", workflowId);
        return ResponseEntity.notFound().build();
      }
      logger.info("Workflow status retrieved: {}", status);
      if (status.getRuntimeStatus().equals(WorkflowRuntimeStatus.COMPLETED)) {
        return ResponseEntity.ok("Workflow ID: " + workflowId + ", Status: " + status);
      }
      return ResponseEntity.accepted().body("Workflow ID: " + workflowId + ", Status: " + status);
    } catch (Exception e) {
      logger.error("Failed to get workflow status: {}", workflowId, e);
      return ResponseEntity.internalServerError().body("Failed to get workflow status: " + e.getMessage());
    }
  }

  @PostMapping("/workflow/terminate/{workflowId}")
  public ResponseEntity<String> terminateWorkflow(@PathVariable String workflowId) {
    try {
      workflowClient.terminateWorkflow(workflowId, null);

      logger.info("Workflow terminated successfully for ID: {}", workflowId);
      return ResponseEntity.ok("Workflow terminated successfully");
    } catch (Exception e) {
      logger.error("Failed to terminate workflow for ID: {}: {}", workflowId, e.getMessage(), e);
      return ResponseEntity.internalServerError().body("Failed to terminate workflow: " + e.getMessage());
    }
  }
}
