import { DateTime } from "luxon";

const relativeTimeDefaultOptions = {
  thresholdDays: 14,
  format: DateTime.DATE_MED_WITH_WEEKDAY,
}

/**
 * Convert the given timestamp to a Luxon time object.
 *
 * @param {Date | string} timestamp either a Date object or an ISO time string.
 * @returns Luxon time object.
 */
function parseTimestamp(timestamp) {
  if (timestamp instanceof Date) {
    return DateTime.fromJSDate(timestamp);
  }
  return DateTime.fromISO(timestamp);
}

// relativeTime parses the timestamp (can be ISO-formatted string or JS Date
// object) and returns it in string form. The returned string is either "xxx
// time ago" if it's a relatively short time ago, or the formatted absolute time
// otherwise.
export function relativeTime(timestamp, options) {
  const parsedTimestamp = parseTimestamp(timestamp);

  if (!options) options = relativeTimeDefaultOptions;

  const now = DateTime.local();
  const ageInDays = now.diff(parsedTimestamp).as('days');
  if (ageInDays > options.format)
    return parsedTimestamp.toLocaleString(options.format);
  return parsedTimestamp.toRelative({style: "narrow"});
}

export function shortened(timestamp) {
  const parsedTimestamp = parseTimestamp(timestamp);
  const now = DateTime.local();
  const ageInHours = now.diff(parsedTimestamp).as('hours');
  if (ageInHours < 24)
    return parsedTimestamp.toLocaleString(DateTime.TIME_24_SIMPLE);
  return parsedTimestamp.toLocaleString(DateTime.DATE_MED_WITH_WEEKDAY);
}
