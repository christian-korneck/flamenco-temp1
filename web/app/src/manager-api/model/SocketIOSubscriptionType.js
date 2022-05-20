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
* Enum class SocketIOSubscriptionType.
* @enum {}
* @readonly
*/
export default class SocketIOSubscriptionType {
    
        /**
         * value: "job"
         * @const
         */
        "job" = "job";

    
        /**
         * value: "tasklog"
         * @const
         */
        "tasklog" = "tasklog";

    

    /**
    * Returns a <code>SocketIOSubscriptionType</code> enum value from a Javascript object name.
    * @param {Object} data The plain JavaScript object containing the name of the enum value.
    * @return {module:model/SocketIOSubscriptionType} The enum <code>SocketIOSubscriptionType</code> value.
    */
    static constructFromObject(object) {
        return object;
    }
}

