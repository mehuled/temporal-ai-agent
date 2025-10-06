package workflows

import (
	"temporal-ai-agent/activities"
	"time"

	"go.temporal.io/sdk/workflow"
)

func SayHelloWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Set up signal channels
	userPromptChan := workflow.GetSignalChannel(ctx, "user_prompt")
	confirmChan := workflow.GetSignalChannel(ctx, "confirm")
	endChatChan := workflow.GetSignalChannel(ctx, "end_chat")

	// Initial greeting
	var result string
	err := workflow.ExecuteActivity(ctx, activities.Greet, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	// Wait for signals in a loop
	for {
		selector := workflow.NewSelector(ctx)
		
		// Add signal channels to selector
		selector.AddReceive(userPromptChan, func(c workflow.ReceiveChannel, more bool) {
			var userMessage string
			c.Receive(ctx, &userMessage)
			workflow.GetLogger(ctx).Info("Received user_prompt signal", "message", userMessage)
			
			// Process user prompt
			var promptResult string
			err := workflow.ExecuteActivity(ctx, activities.Greet, userMessage).Get(ctx, &promptResult)
			if err != nil {
				workflow.GetLogger(ctx).Error("Error processing user prompt", "error", err)
			} else {
				result = promptResult
			}
		})
		
		selector.AddReceive(confirmChan, func(c workflow.ReceiveChannel, more bool) {
			var confirmMessage string
			c.Receive(ctx, &confirmMessage)
			workflow.GetLogger(ctx).Info("Received confirm signal", "message", confirmMessage)
			
			// Process confirmation
			var confirmResult string
			err := workflow.ExecuteActivity(ctx, activities.Greet, "Confirmed: "+confirmMessage).Get(ctx, &confirmResult)
			if err != nil {
				workflow.GetLogger(ctx).Error("Error processing confirmation", "error", err)
			} else {
				result = confirmResult
			}
		})
		
		selector.AddReceive(endChatChan, func(c workflow.ReceiveChannel, more bool) {
			var endMessage string
			c.Receive(ctx, &endMessage)
			workflow.GetLogger(ctx).Info("Received end_chat signal", "message", endMessage)
			
			// End the workflow
			result = "Chat ended: " + endMessage
			return
		})
		
		// Wait for any signal
		selector.Select(ctx)
		
		// Check if we should end the workflow
		if endChatChan.ReceiveAsync(&result) {
			break
		}
	}

	return result, nil
}
