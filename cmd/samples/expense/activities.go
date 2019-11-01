package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.temporal.io/temporal/activity"
	"go.uber.org/zap"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.Register(createExpenseActivity)
	activity.Register(waitForDecisionActivity)
	activity.Register(paymentActivity)
}

func createExpenseActivity(ctx context.Context, expenseID string) error {
	if len(expenseID) == 0 {
		return errors.New("expense id is empty")
	}

	resp, err := http.Get(expenseServerHostPort + "/create?is_api_call=true&id=" + expenseID)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("Expense created.", zap.String("ExpenseID", expenseID))
		return nil
	}

	return errors.New(string(body))
}

// waitForDecisionActivity waits for the expense decision. This activity will complete asynchronously. When this method
// returns error activity.ErrResultPending, the cadence client recognize this error, and won't mark this activity
// as failed or completed. The cadence server will wait until Client.CompleteActivity() is called or timeout happened
// whichever happen first. In this sample case, the CompleteActivity() method is called by our dummy expense server when
// the expense is approved.
func waitForDecisionActivity(ctx context.Context, expenseID string) (string, error) {
	if len(expenseID) == 0 {
		return "", errors.New("expense id is empty")
	}

	logger := activity.GetLogger(ctx)

	// save current activity info so it can be completed asynchronously when expense is approved/rejected
	activityInfo := activity.GetInfo(ctx)
	formData := url.Values{}
	formData.Add("task_token", string(activityInfo.TaskToken))

	registerCallbackURL := expenseServerHostPort + "/registerCallback?id=" + expenseID
	resp, err := http.PostForm(registerCallbackURL, formData)
	if err != nil {
		logger.Info("waitForDecisionActivity failed to register callback.", zap.Error(err))
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	status := string(body)
	if status == "SUCCEED" {
		// register callback succeed
		logger.Info("Successfully registered callback.", zap.String("ExpenseID", expenseID))

		// ErrActivityResultPending is returned from activity's execution to indicate the activity is not completed when it returns.
		// activity will be completed asynchronously when Client.CompleteActivity() is called.
		return "", activity.ErrResultPending
	}

	logger.Warn("Register callback failed.", zap.String("ExpenseStatus", status))
	return "", fmt.Errorf("register callback failed status:%s", status)
}

func paymentActivity(ctx context.Context, expenseID string) error {
	if len(expenseID) == 0 {
		return errors.New("expense id is empty")
	}

	resp, err := http.Get(expenseServerHostPort + "/action?is_api_call=true&type=payment&id=" + expenseID)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("paymentActivity succeed", zap.String("ExpenseID", expenseID))
		return nil
	}

	return errors.New(string(body))
}
