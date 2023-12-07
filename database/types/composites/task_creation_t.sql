DROP TYPE IF EXISTS "task_creation_t";

CREATE TYPE "task_creation_t" AS
(
  "title"       VARCHAR(100),
  "headline"    VARCHAR(100),
  "description" TEXT,
  "priority"    task_priority_t,
  "status"      task_status_t,
  "due_date"    TIMESTAMPTZ,
  "remind_at"   TIMESTAMPTZ
);

ALTER TYPE "task_creation_t"
  OWNER TO "noda";

COMMENT ON TYPE "task_creation_t"
             IS 'Represents the specifications for creating a new task.';
