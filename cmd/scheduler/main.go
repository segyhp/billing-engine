package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segyhp/billing-engine/internal/config"

	"github.com/robfig/cron/v3"
)

func main() {
	log.Println("Starting billing scheduler...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize cron scheduler
	c := cron.New(cron.WithSeconds())

	// Schedule tasks
	setupCronJobs(c, cfg)

	// Start the scheduler
	c.Start()
	log.Println("Scheduler started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down scheduler...")
	c.Stop()
	log.Println("Scheduler stopped")
}

func setupCronJobs(c *cron.Cron, cfg *config.Config) {
	// Daily job to update overdue payments (runs at midnight)
	_, err := c.AddFunc("0 0 0 * * *", func() {
		log.Println("Running daily overdue payment update job...")
		// TODO: Implement overdue payment update logic
		updateOverduePayments()
	})
	if err != nil {
		log.Printf("Error scheduling overdue payment update job: %v", err)
	}

	// Weekly job to send payment reminders (runs on Sundays at 9 AM)
	_, err = c.AddFunc("0 0 9 * * SUN", func() {
		log.Println("Running weekly payment reminder job...")
		// TODO: Implement payment reminder logic
		sendPaymentReminders()
	})
	if err != nil {
		log.Printf("Error scheduling payment reminder job: %v", err)
	}

	log.Println("Cron jobs scheduled successfully")
}

// TODO: Implement this function to mark overdue payments
func updateOverduePayments() {
	// Business logic to implement:
	// 1. Get all active loans
	// 2. For each loan, check which payments are overdue
	// 3. Update loan_schedule status from 'pending' to 'overdue'
	// 4. Update loan status to 'delinquent' if applicable
	log.Println("TODO: Implement updateOverduePayments logic")
}

// TODO: Implement this function to send payment reminders
func sendPaymentReminders() {
	// Business logic to implement:
	// 1. Get all loans with upcoming payments (due in next 3 days)
	// 2. Send notification/reminder to borrowers
	// 3. Log reminder sent
	log.Println("TODO: Implement sendPaymentReminders logic")
}
