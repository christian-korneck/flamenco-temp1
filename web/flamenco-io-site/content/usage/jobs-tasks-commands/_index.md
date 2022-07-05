---
title: Jobs, Tasks, and Commands
---

TODO: write about the pipeline from job submission to command execution.


## Task Statuses

The following table shows the meaning of the different task statuses:

| Status        | Meaning | Possible next status |
| ------------- | ------- | ----------- |
| `queued`      | Ready to be worked on by a Worker | `active`, `canceled` |
| `active`      | Assigned to a worker for execution | `completed`, `canceled`, `failed`, `soft-failed` |
| `soft-failed` | Same as `queued`, but has been failed by a worker in an earlier execution | `completed`, `failed`, `canceled` |
| `completed`   | Worker executed the task succesfully | `requeued` |
| `paused`      | Not yet implemented | |
