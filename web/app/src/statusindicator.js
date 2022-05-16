import { toTitleCase } from '@/strings';

/**
 * Construct HTML for a certain job/task status.
 *
 * This function is implemented here as JavaScript, because we don't know how to
 * get Tabulator to use Vue components in cells.
 *
 * @param {string} status The job/task status. Assumed to only consist of
 *     letters and dashes, HTML-safe, and valid as a CSS class name.
 * @returns the HTML for the status indicator.
 */
export function indicator(status) {
  const label = toTitleCase(status);
  return `<span title="${label}" class="indicator status-${status}"></span>`;
}
