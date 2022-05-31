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
import SocketIOSubscriptionOperation from './SocketIOSubscriptionOperation';
import SocketIOSubscriptionType from './SocketIOSubscriptionType';

/**
 * The SocketIOSubscription model module.
 * @module model/SocketIOSubscription
 * @version 0.0.0
 */
class SocketIOSubscription {
    /**
     * Constructs a new <code>SocketIOSubscription</code>.
     * Send by SocketIO clients as &#x60;/subscription&#x60; event type, to manage their subscription to job updates. Clients always get job updates, but for task updates or task logs they need to explicitly subscribe. For simplicity, clients can only subscribe to one job (to get task updates for that job) and one task&#39;s log at a time. 
     * @alias module:model/SocketIOSubscription
     * @param op {module:model/SocketIOSubscriptionOperation} 
     * @param type {module:model/SocketIOSubscriptionType} 
     */
    constructor(op, type) { 
        
        SocketIOSubscription.initialize(this, op, type);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, op, type) { 
        obj['op'] = op;
        obj['type'] = type;
    }

    /**
     * Constructs a <code>SocketIOSubscription</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/SocketIOSubscription} obj Optional instance to populate.
     * @return {module:model/SocketIOSubscription} The populated <code>SocketIOSubscription</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new SocketIOSubscription();

            if (data.hasOwnProperty('op')) {
                obj['op'] = SocketIOSubscriptionOperation.constructFromObject(data['op']);
            }
            if (data.hasOwnProperty('type')) {
                obj['type'] = SocketIOSubscriptionType.constructFromObject(data['type']);
            }
            if (data.hasOwnProperty('uuid')) {
                obj['uuid'] = ApiClient.convertToType(data['uuid'], 'String');
            }
        }
        return obj;
    }


}

/**
 * @member {module:model/SocketIOSubscriptionOperation} op
 */
SocketIOSubscription.prototype['op'] = undefined;

/**
 * @member {module:model/SocketIOSubscriptionType} type
 */
SocketIOSubscription.prototype['type'] = undefined;

/**
 * UUID of the thing to subscribe to / unsubscribe from.
 * @member {String} uuid
 */
SocketIOSubscription.prototype['uuid'] = undefined;






export default SocketIOSubscription;

