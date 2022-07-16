/**
 * Flamenco manager
 * Render Farm manager API
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 *
 */

import ApiClient from '../ApiClient';

/**
 * The WorkerSleepSchedule model module.
 * @module model/WorkerSleepSchedule
 * @version 0.0.0
 */
class WorkerSleepSchedule {
    /**
     * Constructs a new <code>WorkerSleepSchedule</code>.
     * Sleep schedule for a single Worker. Start and end time indicate the time of each day at which the schedule is active. Applies only when today is in &#x60;days_of_week&#x60;, or when &#x60;days_of_week&#x60; is empty. Start and end time are in 24-hour HH:MM notation. 
     * @alias module:model/WorkerSleepSchedule
     * @param isActive {Boolean} 
     * @param daysOfWeek {String} Space-separated two-letter strings indicating days of week the schedule is active (\"mo\", \"tu\", etc.). Empty means \"every day\". 
     * @param startTime {String} 
     * @param endTime {String} 
     */
    constructor(isActive, daysOfWeek, startTime, endTime) { 
        
        WorkerSleepSchedule.initialize(this, isActive, daysOfWeek, startTime, endTime);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, isActive, daysOfWeek, startTime, endTime) { 
        obj['is_active'] = isActive;
        obj['days_of_week'] = daysOfWeek;
        obj['start_time'] = startTime;
        obj['end_time'] = endTime;
    }

    /**
     * Constructs a <code>WorkerSleepSchedule</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/WorkerSleepSchedule} obj Optional instance to populate.
     * @return {module:model/WorkerSleepSchedule} The populated <code>WorkerSleepSchedule</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new WorkerSleepSchedule();

            if (data.hasOwnProperty('is_active')) {
                obj['is_active'] = ApiClient.convertToType(data['is_active'], 'Boolean');
            }
            if (data.hasOwnProperty('days_of_week')) {
                obj['days_of_week'] = ApiClient.convertToType(data['days_of_week'], 'String');
            }
            if (data.hasOwnProperty('start_time')) {
                obj['start_time'] = ApiClient.convertToType(data['start_time'], 'String');
            }
            if (data.hasOwnProperty('end_time')) {
                obj['end_time'] = ApiClient.convertToType(data['end_time'], 'String');
            }
        }
        return obj;
    }


}

/**
 * @member {Boolean} is_active
 */
WorkerSleepSchedule.prototype['is_active'] = undefined;

/**
 * Space-separated two-letter strings indicating days of week the schedule is active (\"mo\", \"tu\", etc.). Empty means \"every day\". 
 * @member {String} days_of_week
 */
WorkerSleepSchedule.prototype['days_of_week'] = undefined;

/**
 * @member {String} start_time
 */
WorkerSleepSchedule.prototype['start_time'] = undefined;

/**
 * @member {String} end_time
 */
WorkerSleepSchedule.prototype['end_time'] = undefined;






export default WorkerSleepSchedule;

