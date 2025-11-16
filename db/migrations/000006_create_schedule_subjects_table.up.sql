-- +migrate Up
CREATE TABLE schedule_subjects (
    schedule_id INTEGER NOT NULL,
    subject_id INTEGER NOT NULL,
    PRIMARY KEY (schedule_id, subject_id),
    FOREIGN KEY (schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id) ON DELETE CASCADE
);
