package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vickey290/go-modules/server/models"
)

const (
	retryPeriod = time.Minute
)

type RemindersMap map[int]map[int]models.Reminder

func (rMap RemindersMap) flatten(id int) (int, models.Reminder) {
	var index int
	var reminder models.Reminder
	for i, r := range rMap[id] {
		index = i
		reminder = r
	}
	return index, reminder
}

type ReminderRepository interface {
	Save([]models.Reminder) (int, error)
	Filter(filterFn func(reminder models.Reminder) bool) (RemindersMap, error)
	NextID() int
}

type Snapshot struct {
	All         RemindersMap
	UnCompleted RemindersMap
}

type Reminders struct {
	repo     ReminderRepository
	Snapshot Snapshot
}

func NewReminders(repo ReminderRepository) *Reminders {
	return &Reminders{
		repo: repo,
		Snapshot: Snapshot{
			All:         RemindersMap{},
			UnCompleted: RemindersMap{},
		},
	}
}

func (s *Reminders) Populate() error {
	all, err := s.repo.Filter(nil)
	if err != nil {
		return models.WrapError("could not get all reminders", err)
	}
	unCompleted, err := s.repo.Filter(func(r models.Reminder) bool {
		return r.ModifiedAt.Add(r.Duration).UnixNano() > time.Now().UnixNano()
	})
	if err != nil {
		return models.WrapError("could not get uncompleted reminders", err)
	}
	s.Snapshot.All = all
	s.Snapshot.UnCompleted = unCompleted
	return nil
}

type ReminderCreateBody struct {
	Title    string
	Message  string
	Duration time.Duration
}

func (s Reminders) Create(body ReminderCreateBody) (models.Reminder, error) {
	nextID := s.repo.NextID()
	if body.Title == "" {
		err := models.DataValidationError{
			Message: "title cannot be empty",
		}
		return models.Reminder{}, err
	}
	if body.Message == "" {
		err := models.DataValidationError{
			Message: "body cannot be empty",
		}
		return models.Reminder{}, err
	}
	if body.Duration == 0 {
		err := models.DataValidationError{
			Message: "duration cannot be 0",
		}
		return models.Reminder{}, err
	}
	reminder := models.Reminder{
		ID:         nextID,
		Title:      body.Title,
		Message:    body.Message,
		Duration:   body.Duration,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	index := len(s.Snapshot.All)
	s.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	s.Snapshot.UnCompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
	return reminder, nil
}

type ReminderEditBody struct {
	ID       int
	Title    string
	Message  string
	Duration time.Duration
}

func (s Reminders) Edit(reminderBody ReminderEditBody) (models.Reminder, error) {
	_, ok := s.Snapshot.All[reminderBody.ID]
	if !ok {
		err := models.NotFoundError{
			Message: fmt.Sprintf("could not find reminder with id: %d", reminderBody.ID),
		}
		return models.Reminder{}, err
	}
	changed := false
	index, reminder := s.Snapshot.All.flatten(reminderBody.ID)
	if strings.TrimSpace(reminderBody.Title) != "" {
		reminder.Title = reminderBody.Title
		changed = true
	}
	if strings.TrimSpace(reminderBody.Message) != "" {
		reminder.Message = reminderBody.Message
		changed = true
	}
	if reminderBody.Duration != 0 {
		reminder.Duration = reminderBody.Duration
		changed = true
	}
	if !changed {
		err := models.FormatValidationError{
			Message: "body must contain at least 1 of: 'title', 'message', 'duration'",
		}
		return models.Reminder{}, err
	}
	reminder.ModifiedAt = time.Now()
	s.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	if reminder.ModifiedAt.UnixNano() < time.Now().Add(reminderBody.Duration).UnixNano() {
		s.Snapshot.UnCompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
	} else {
		delete(s.Snapshot.UnCompleted, reminder.ID)
	}
	return reminder, nil
}

func (s Reminders) Fetch(ids []int) ([]models.Reminder, error) {
	reminders := make([]models.Reminder, 0)
	var notFound []int
	for _, id := range ids {
		_, ok := s.Snapshot.All[id]
		if !ok {
			notFound = append(notFound, id)
			continue
		}
		_, reminder := s.Snapshot.All.flatten(id)
		reminders = append(reminders, reminder)
	}
	if len(notFound) > 0 {
		err := models.NotFoundError{
			Message: fmt.Sprintf("could not find reminders with ids: %v", notFound),
		}
		return []models.Reminder{}, err
	}
	return reminders, nil
}

// Delete deletes a list of reminders and persists the changes
func (s Reminders) Delete(ids []int) error {
	var notFound []int
	for _, id := range ids {
		_, ok := s.Snapshot.All[id]
		if !ok {
			notFound = append(notFound, id)
		}
	}
	if len(notFound) > 0 {
		return models.NotFoundError{
			Message: fmt.Sprintf("could not find reminders with ids: %v", notFound),
		}
	}

	for _, id := range ids {
		delete(s.Snapshot.All, id)
		delete(s.Snapshot.UnCompleted, id)
	}
	return nil
}

// save saves the current reminders snapshot
func (s Reminders) save() error {
	reminders := make([]models.Reminder, len(s.Snapshot.All))
	for _, reminderMap := range s.Snapshot.All {
		for i, reminder := range reminderMap {
			reminders[i] = reminder
		}
	}

	n, err := s.repo.Save(reminders)
	if err != nil {
		return models.WrapError("could not save snapshot", err)
	}
	if n > 0 && len(reminders) != 0 {
		log.Printf("successfully saved snapshot: %d reminders", len(reminders))
	}
	return nil
}

// GetSnapshot fetches the current service snapshot
func (s Reminders) snapshot() Snapshot {
	return s.Snapshot
}

// snapshotGrooming clears the current snapshot from notified reminders
func (s Reminders) snapshotGrooming(notifiedReminders ...models.Reminder) {
	if len(notifiedReminders) > 0 {
		log.Printf("snapshot grooming: %d record(s)", len(notifiedReminders))
	}
	for _, reminder := range notifiedReminders {
		delete(s.Snapshot.UnCompleted, reminder.ID)
		reminder.Duration = -time.Hour
		index, _ := s.Snapshot.All.flatten(reminder.ID)
		s.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	}
}

func (s Reminders) retry(reminder models.Reminder, d time.Duration) {
	reminder.ModifiedAt = time.Now()
	if d <= 0 {
		reminder.Duration = retryPeriod
	} else {
		reminder.Duration = d
	}
	log.Printf(
		"retrying record with id: %d after %v",
		reminder.ID,
		reminder.Duration.String(),
	)
	index, _ := s.Snapshot.All.flatten(reminder.ID)
	s.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	s.Snapshot.UnCompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
}
