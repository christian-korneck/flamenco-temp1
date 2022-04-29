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
import TaskStatus from './TaskStatus';

/**
 * The TaskSummary model module.
 * @module model/TaskSummary
 * @version 0.0.0
 */
class TaskSummary {
    /**
     * Constructs a new <code>TaskSummary</code>.
     * Just enough information about the task to show in the job&#39;s task list.
     * @alias module:model/TaskSummary
     * @param id {String} 
     * @param name {String} 
     * @param status {module:model/TaskStatus} 
     * @param priority {Number} 
     * @param taskType {String} 
     * @param updated {Date} 
     */
    constructor(id, name, status, priority, taskType, updated) { 
        
        TaskSummary.initialize(this, id, name, status, priority, taskType, updated);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, id, name, status, priority, taskType, updated) { 
        obj['id'] = id;
        obj['name'] = name;
        obj['status'] = status;
        obj['priority'] = priority;
        obj['task_type'] = taskType;
        obj['updated'] = updated;
    }

    /**
     * Constructs a <code>TaskSummary</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/TaskSummary} obj Optional instance to populate.
     * @return {module:model/TaskSummary} The populated <code>TaskSummary</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new TaskSummary();

            if (data.hasOwnProperty('id')) {
                obj['id'] = ApiClient.convertToType(data['id'], 'String');
            }
            if (data.hasOwnProperty('name')) {
                obj['name'] = ApiClient.convertToType(data['name'], 'String');
            }
            if (data.hasOwnProperty('status')) {
                obj['status'] = TaskStatus.constructFromObject(data['status']);
            }
            if (data.hasOwnProperty('priority')) {
                obj['priority'] = ApiClient.convertToType(data['priority'], 'Number');
            }
            if (data.hasOwnProperty('task_type')) {
                obj['task_type'] = ApiClient.convertToType(data['task_type'], 'String');
            }
            if (data.hasOwnProperty('updated')) {
                obj['updated'] = ApiClient.convertToType(data['updated'], 'Date');
            }
        }
        return obj;
    }


}

/**
 * @member {String} id
 */
TaskSummary.prototype['id'] = undefined;

/**
 * @member {String} name
 */
TaskSummary.prototype['name'] = undefined;

/**
 * @member {module:model/TaskStatus} status
 */
TaskSummary.prototype['status'] = undefined;

/**
 * @member {Number} priority
 */
TaskSummary.prototype['priority'] = undefined;

/**
 * @member {String} task_type
 */
TaskSummary.prototype['task_type'] = undefined;

/**
 * @member {Date} updated
 */
TaskSummary.prototype['updated'] = undefined;






export default TaskSummary;
