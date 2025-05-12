-- Insert dummy students
INSERT INTO students (name, latitude, longitude, token)
VALUES 
  ('Alice Johnson', 27.7000, 85.3333, 123456),
  ('Bob Smith', 27.7050, 85.3200, 234567),
  ('Charlie Brown', 27.7100, 85.3150, 345678);

-- Insert a classroom
INSERT INTO classrooms (name)
VALUES ('Computer Science 101');

-- Get classroom ID
-- (Assuming ID is 1, adjust if needed)

-- Enroll students in the classroom
INSERT INTO student_classroom_enrollment (student_id, classroom_id)
VALUES 
  (1, 1),
  (2, 1),
  (3, 1);

-- Insert a class session
INSERT INTO classes (classroom_id, start_time, end_time, link)
VALUES 
  (1, NOW() + INTERVAL '1 hour', NOW() + INTERVAL '2 hours', 'https://meet.google.com/example');

-- Record attendance
-- Assume class_id = 1
INSERT INTO attendance (student_id, class_id, status)
VALUES 
  (1, 1, 'present'),
  (2, 1, 'absent'),
  (3, 1, 'present');

-- Record summary stats
INSERT INTO record (student_id, classroom_id, present_count, absent_count, last_attended)
VALUES 
  (1, 1, 1, 0, NOW()),
  (2, 1, 0, 1, NOW()),
  (3, 1, 1, 0, NOW());
