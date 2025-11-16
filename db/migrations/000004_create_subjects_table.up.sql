-- +migrate Up
CREATE TABLE subjects (
    subject_id INTEGER PRIMARY KEY AUTOINCREMENT,
    career_id INTEGER,
    department TEXT,
    subject_name TEXT,
    semester INTEGER,
    section TEXT,
    teacher_title TEXT,
    teacher_lastname TEXT,
    teacher_name TEXT,
    teacher_email TEXT,

    -- Weekly schedule
    monday TEXT,
    monday_classroom TEXT,
    tuesday TEXT,
    tuesday_classroom TEXT,
    wednesday TEXT,
    wednesday_classroom TEXT,
    thursday TEXT,
    thursday_classroom TEXT,
    friday TEXT,
    friday_classroom TEXT,
    saturday TEXT,
    saturday_night_dates TEXT,

    -- Exams
    partial1_date DATE,
    partial1_time TEXT,
    partial1_classroom TEXT,
    partial2_date DATE,
    partial2_time TEXT,
    partial2_classroom TEXT,
    final1_date DATE,
    final1_time TEXT,
    final1_classroom TEXT,
    final1_review_date DATE,
    final1_review_time TEXT,
    final2_date DATE,
    final2_time TEXT,
    final2_classroom TEXT,
    final2_review_date DATE,
    final2_review_time TEXT,

    -- Committee
    committee_chair TEXT,
    committee_member1 TEXT,
    committee_member2 TEXT,

    FOREIGN KEY (career_id) REFERENCES careers(career_id) ON DELETE SET NULL
);

CREATE INDEX idx_subjects_name ON subjects(subject_name);
