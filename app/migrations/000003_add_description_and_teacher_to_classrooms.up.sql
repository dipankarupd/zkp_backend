-- up.sql
ALTER TABLE classrooms
ADD COLUMN description TEXT,
ADD COLUMN teacher_name VARCHAR(255);

UPDATE classrooms SET description = 'Intro to CS', teacher_name = 'Prof. Alan' WHERE id = 1;
UPDATE classrooms SET description = 'Basic Math Concepts', teacher_name = 'Dr. Smith' WHERE id = 2;
