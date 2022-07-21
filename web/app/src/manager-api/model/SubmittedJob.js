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
 * The SubmittedJob model module.
 * @module model/SubmittedJob
 * @version 0.0.0
 */
class SubmittedJob {
    /**
     * Constructs a new <code>SubmittedJob</code>.
     * Job definition submitted to Flamenco.
     * @alias module:model/SubmittedJob
     * @param name {String} 
     * @param type {String} 
     * @param priority {Number} 
     * @param submitterPlatform {String} Operating system of the submitter. This is used to recognise two-way variables. This should be a lower-case version of the platform, like \"linux\", \"windows\", \"darwin\", \"openbsd\", etc. Should be ompatible with Go's `runtime.GOOS`; run `go tool dist list` to get a list of possible platforms. As a special case, the platform \"manager\" can be given, which will be interpreted as \"the Manager's platform\". This is mostly to make test/debug scripts easier, as they can use a static document on all platforms. 
     */
    constructor(name, type, priority, submitterPlatform) { 
        
        SubmittedJob.initialize(this, name, type, priority, submitterPlatform);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, name, type, priority, submitterPlatform) { 
        obj['name'] = name;
        obj['type'] = type;
        obj['priority'] = priority || 50;
        obj['submitter_platform'] = submitterPlatform;
    }

    /**
     * Constructs a <code>SubmittedJob</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/SubmittedJob} obj Optional instance to populate.
     * @return {module:model/SubmittedJob} The populated <code>SubmittedJob</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new SubmittedJob();

            if (data.hasOwnProperty('name')) {
                obj['name'] = ApiClient.convertToType(data['name'], 'String');
            }
            if (data.hasOwnProperty('type')) {
                obj['type'] = ApiClient.convertToType(data['type'], 'String');
            }
            if (data.hasOwnProperty('priority')) {
                obj['priority'] = ApiClient.convertToType(data['priority'], 'Number');
            }
            if (data.hasOwnProperty('settings')) {
                obj['settings'] = ApiClient.convertToType(data['settings'], {'String': Object});
            }
            if (data.hasOwnProperty('metadata')) {
                obj['metadata'] = ApiClient.convertToType(data['metadata'], {'String': 'String'});
            }
            if (data.hasOwnProperty('submitter_platform')) {
                obj['submitter_platform'] = ApiClient.convertToType(data['submitter_platform'], 'String');
            }
        }
        return obj;
    }


}

/**
 * @member {String} name
 */
SubmittedJob.prototype['name'] = undefined;

/**
 * @member {String} type
 */
SubmittedJob.prototype['type'] = undefined;

/**
 * @member {Number} priority
 * @default 50
 */
SubmittedJob.prototype['priority'] = 50;

/**
 * @member {Object.<String, Object>} settings
 */
SubmittedJob.prototype['settings'] = undefined;

/**
 * Arbitrary metadata strings. More complex structures can be modeled by using `a.b.c` notation for the key.
 * @member {Object.<String, String>} metadata
 */
SubmittedJob.prototype['metadata'] = undefined;

/**
 * Operating system of the submitter. This is used to recognise two-way variables. This should be a lower-case version of the platform, like \"linux\", \"windows\", \"darwin\", \"openbsd\", etc. Should be ompatible with Go's `runtime.GOOS`; run `go tool dist list` to get a list of possible platforms. As a special case, the platform \"manager\" can be given, which will be interpreted as \"the Manager's platform\". This is mostly to make test/debug scripts easier, as they can use a static document on all platforms. 
 * @member {String} submitter_platform
 */
SubmittedJob.prototype['submitter_platform'] = undefined;






export default SubmittedJob;

