import { DateTime } from "luxon";

const relativeTimeDefaultOptions = {
  thresholdDays: 14,
  format: DateTime.DATE_MED_WITH_WEEKDAY,
}

// relativeTime parses the timestamp (can be ISO-formatted string or JS Date
// object) and returns it in string form. The returned string is either "xxx
// time ago" if it's a relatively short time ago, or the formatted absolute time
// otherwise.
export function relativeTime(timestamp, options) {
  let parsedTimestamp = null;
  if (timestamp instanceof Date) {
    parsedTimestamp = DateTime.fromJSDate(timestamp);
  } else {
    parsedTimestamp = DateTime.fromISO(timestamp);
  }

  if (!options) options = relativeTimeDefaultOptions;

  const now = DateTime.local();
  const ageInDays = now.diff(parsedTimestamp).as('days');
  if (ageInDays > options.format)
    return parsedTimestamp.toLocaleString(options.format);
  return parsedTimestamp.toRelative({style: "narrow"});
}
