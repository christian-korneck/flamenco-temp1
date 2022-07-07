import { toTitleCase } from '@/strings';

/**
 * Construct HTML for a certain job/task status.
 *
 * This function is implemented here as JavaScript, because we don't know how to
 * get Tabulator to use Vue components in cells.
 *
 * @param {string} status The job/task status. Assumed to only consist of
 *     letters and dashes, HTML-safe, and valid as a CSS class name.
 * @param {string} classNamePrefix optional prefix used for the class name
 * @returns the HTML for the status indicator.
 */
export function indicator(status, classNamePrefix) {
  const label = toTitleCase(status);
  if (!classNamePrefix) classNamePrefix = ""; // force an empty string for any false value.
  return `<span title="${label}" class="indicator ${classNamePrefix}status-${status}"></span>`;
}

/**
 * Construct HTML for showing a worker's status, including any status change
 * request.
 *
 * @param {API.WorkerSummary} workerInfo
 * @returns the HTML for the worker status.
 */
export function workerStatus(worker) {
  if (!worker.status_change) {
    return `<span class="worker-status-${worker.status}">${worker.status}</span>`;
  }

  let arrow;
  if (worker.status_change.is_lazy) {
    arrow = `<span class='state-transition-arrow lazy' title='lazy status transition'>➠</span>`
  } else {
    arrow = `<span class='state-transition-arrow forced' title='forced status transition'>➜</span>`
  }

  return `<span class="worker-status-${worker.status}">${worker.status}</span>
          ${arrow}
          <span class="worker-status-${worker.status_change.status}">${worker.status_change.status}</span>`;
}
