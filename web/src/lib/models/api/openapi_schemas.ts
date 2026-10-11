/**
 * @swagger
 * components:
 *   schemas:
 *     APIToken:
 *       type: object
 *       required:
 *         - id
 *         - name
 *         - user
 *       properties:
 *         id:
 *           type: string
 *           description: Token ID (15 chars)
 *         name:
 *           type: string
 *           description: Token name
 *         user:
 *           type: string
 *           description: User ID that owns this token
 *         expiration:
 *           type: string
 *           format: date-time
 *           description: Optional expiration date
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     APITokenInput:
 *       type: object
 *       required:
 *         - name
 *         - user
 *       properties:
 *         name:
 *           type: string
 *           minLength: 1
 *           description: Token name
 *         expiration:
 *           type: string
 *           format: date-time
 *           description: Optional expiration date (must be in future)
 *         user:
 *           type: string
 *           description: User ID (15 chars)
 *
 *     User:
 *       type: object
 *       required:
 *         - id
 *         - username
 *         - email
 *       properties:
 *         id:
 *           type: string
 *           description: User ID (15 chars)
 *         username:
 *           type: string
 *           description: Username (3+ chars, alphanumeric with dots)
 *         email:
 *           type: string
 *           format: email
 *           description: User email address
 *         avatar:
 *           type: string
 *           description: Avatar file path
 *         verified:
 *           type: boolean
 *           description: Email verification status
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     UserCreateInput:
 *       type: object
 *       required:
 *         - username
 *         - email
 *         - password
 *       properties:
 *         username:
 *           type: string
 *           minLength: 3
 *           pattern: '^\w[\w\.]*$'
 *           description: Username (3+ chars, starts with word char)
 *         email:
 *           type: string
 *           format: email
 *           description: User email address
 *         password:
 *           type: string
 *           minLength: 8
 *           maxLength: 72
 *           description: Password (8-72 chars)
 *         passwordConfirm:
 *           type: string
 *           description: Password confirmation (must match password)
 *
 *     UserUpdateInput:
 *       type: object
 *       properties:
 *         username:
 *           type: string
 *           minLength: 3
 *           pattern: '^\w[\w\.]*$'
 *         email:
 *           type: string
 *           format: email
 *         password:
 *           type: string
 *           minLength: 8
 *           maxLength: 72
 *         oldPassword:
 *           type: string
 *           description: Required when changing password
 *         passwordConfirm:
 *           type: string
 *           description: Must match password
 *
 *     Tag:
 *       type: object
 *       required:
 *         - id
 *         - name
 *       properties:
 *         id:
 *           type: string
 *           description: Tag ID (15 chars)
 *         name:
 *           type: string
 *           description: Tag name
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     TagInput:
 *       type: object
 *       required:
 *         - name
 *       properties:
 *         name:
 *           type: string
 *           maxLength: 5000
 *           pattern: '^[^\u0000-\u001f\u007f]*$(?![\s\S])'
 *           description: Plain-text tag name. Empty names, Unicode, punctuation, and literal markup are preserved. C0 and DEL controls are rejected. Length is measured in Unicode code points.
 *
 *     TagUpdateInput:
 *       type: object
 *       properties:
 *         name:
 *           type: string
 *           maxLength: 5000
 *           pattern: '^[^\u0000-\u001f\u007f]*$(?![\s\S])'
 *           description: Plain-text tag name. Empty names, Unicode, punctuation, and literal markup are preserved. C0 and DEL controls are rejected. Length is measured in Unicode code points.
 *
 *     CategoryTranslation:
 *       type: object
 *       properties:
 *         name:
 *           type: string
 *         short_name:
 *           type: string
 *
 *     Category:
 *       type: object
 *       required:
 *         - id
 *         - name
 *       properties:
 *         id:
 *           type: string
 *           description: Category ID (15 chars)
 *         name:
 *           type: string
 *         short_name:
 *           type: string
 *           nullable: true
 *         icon:
 *           type: string
 *           nullable: true
 *         translations:
 *           type: object
 *           nullable: true
 *           additionalProperties:
 *             $ref: '#/components/schemas/CategoryTranslation'
 *         settings:
 *           type: object
 *           nullable: true
 *           properties:
 *             wp_merge_enabled:
 *               type: boolean
 *             wp_merge_radius:
 *               type: number
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     Subcategory:
 *       type: object
 *       required:
 *         - id
 *         - category
 *         - name
 *       properties:
 *         id:
 *           type: string
 *           description: Subcategory ID (15 chars)
 *         category:
 *           type: string
 *           description: Parent category ID (15 chars)
 *         name:
 *           type: string
 *         short_name:
 *           type: string
 *           nullable: true
 *         icon:
 *           type: string
 *           nullable: true
 *         badge_icon:
 *           type: string
 *           nullable: true
 *         translations:
 *           type: object
 *           nullable: true
 *           additionalProperties:
 *             $ref: '#/components/schemas/CategoryTranslation'
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     UserCategoryPreference:
 *       type: object
 *       required:
 *         - id
 *         - user
 *         - category
 *         - visible
 *       properties:
 *         id:
 *           type: string
 *           description: Preference ID (15 chars)
 *         user:
 *           type: string
 *           description: User ID (15 chars)
 *         category:
 *           type: string
 *           description: Category ID (15 chars)
 *         visible:
 *           type: boolean
 *         priority:
 *           type: integer
 *           nullable: true
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     UserCategoryPreferenceUpsertInput:
 *       type: object
 *       required:
 *         - category
 *         - visible
 *       properties:
 *         category:
 *           type: string
 *           description: Category ID (15 chars)
 *         visible:
 *           type: boolean
 *
 *     UserCategoryPreferenceReorderInput:
 *       type: object
 *       required:
 *         - categories
 *       properties:
 *         categories:
 *           type: array
 *           items:
 *             type: string
 *             description: Category ID (15 chars)
 *
 *     UserSubcategoryPreference:
 *       type: object
 *       required:
 *         - id
 *         - user
 *         - subcategory
 *         - visible
 *       properties:
 *         id:
 *           type: string
 *           description: Preference ID (15 chars)
 *         user:
 *           type: string
 *           description: User ID (15 chars)
 *         subcategory:
 *           type: string
 *           description: Subcategory ID (15 chars)
 *         visible:
 *           type: boolean
 *         priority:
 *           type: integer
 *           nullable: true
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     UserSubcategoryPreferenceUpsertInput:
 *       type: object
 *       required:
 *         - subcategory
 *         - visible
 *       properties:
 *         subcategory:
 *           type: string
 *           description: Subcategory ID (15 chars)
 *         visible:
 *           type: boolean
 *
 *     UserSubcategoryPreferenceReorderInput:
 *       type: object
 *       required:
 *         - category
 *         - subcategories
 *       properties:
 *         category:
 *           type: string
 *           description: Category ID (15 chars)
 *         subcategories:
 *           type: array
 *           items:
 *             type: string
 *             description: Subcategory ID (15 chars)
 *
 *     Trail:
 *       type: object
 *       required:
 *         - id
 *         - name
 *         - author
 *         - public
 *         - completed
 *       properties:
 *         id:
 *           type: string
 *           description: Trail ID (15 chars)
 *         name:
 *           type: string
 *           description: Trail name
 *         description:
 *           type: string
 *         location:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID
 *         date:
 *           type: string
 *           format: date
 *         public:
 *           type: boolean
 *         completed:
 *           type: boolean
 *         completed_at:
 *           type: string
 *           format: date-time
 *         difficulty:
 *           type: string
 *           enum: [easy, moderate, difficult]
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance:
 *           type: number
 *           description: Distance in meters
 *         elevation_gain:
 *           type: number
 *           description: Elevation gain in meters
 *         elevation_loss:
 *           type: number
 *           description: Elevation loss in meters
 *         duration:
 *           type: number
 *           description: Duration in seconds
 *         photos:
 *           type: array
 *           readOnly: true
 *           description: >
 *             Derived from the photo assets linked to the trail and not writable. Each entry is a file URL,
 *             `/api/v1/files/{assetsCollectionId}/{assetId}/{file}` for a stored photo or
 *             `/api/v1/assets/{assetId}/file` for a photo kept in a connected photo library.
 *             Manage photos through `PUT /api/v1/assets` and `DELETE /api/v1/assets/{id}`, and choose the
 *             thumbnail through `PATCH /api/v1/trail/{id}/thumbnail`.
 *           items:
 *             type: string
 *         thumbnail:
 *           type: integer
 *           description: >
 *             Index of the thumbnail in `photos`: the photo marked as thumbnail through
 *             `PATCH /api/v1/trail/{id}/thumbnail` (`trail_assets.is_thumbnail`), or else the first photo. A generated
 *             route preview counts only when the trail has no other photo.
 *         like_count:
 *           type: integer
 *           default: 0
 *         category:
 *           type: string
 *           description: Category ID (15 chars)
 *         subcategory:
 *           type: string
 *           description: Subcategory ID (15 chars)
 *         tags:
 *           type: array
 *           items:
 *             type: string
 *         gpx:
 *           type: string
 *           description: GPX file path
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     TrailCreateInput:
 *       type: object
 *       required:
 *         - name
 *         - author
 *         - public
 *         - completed
 *       properties:
 *         id:
 *           type: string
 *           description: Optional custom ID (15 chars)
 *         name:
 *           type: string
 *           minLength: 1
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         description:
 *           type: string
 *         location:
 *           type: string
 *         date:
 *           type: string
 *           format: date
 *         public:
 *           type: boolean
 *         completed:
 *           type: boolean
 *         completed_at:
 *           type: string
 *           format: date-time
 *         difficulty:
 *           type: string
 *           enum: [easy, moderate, difficult]
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance:
 *           type: number
 *         elevation_gain:
 *           type: number
 *         elevation_loss:
 *           type: number
 *         duration:
 *           type: number
 *         thumbnail:
 *           type: integer
 *           description: >
 *             Legacy index of the thumbnail in the trail's `photos`. Trail responses use the photo marked through
 *             `PATCH /api/v1/trail/{id}/thumbnail` (`trail_assets.is_thumbnail`) or else the first photo, not this index.
 *         like_count:
 *           type: integer
 *           default: 0
 *         category:
 *           type: string
 *         subcategory:
 *           type: string
 *         tags:
 *           type: array
 *           items:
 *             type: string
 *           default: []
 *         gpx:
 *           type: string
 *
 *     TrailUpdateInput:
 *       type: object
 *       properties:
 *         name:
 *           type: string
 *         description:
 *           type: string
 *         location:
 *           type: string
 *         date:
 *           type: string
 *           format: date
 *         public:
 *           type: boolean
 *         completed:
 *           type: boolean
 *         completed_at:
 *           type: string
 *           format: date-time
 *         difficulty:
 *           type: string
 *           enum: [easy, moderate, difficult]
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance:
 *           type: number
 *         elevation_gain:
 *           type: number
 *         elevation_loss:
 *           type: number
 *         duration:
 *           type: number
 *         thumbnail:
 *           type: integer
 *           description: >
 *             Legacy index of the thumbnail in the trail's `photos`. Trail responses use the photo marked through
 *             `PATCH /api/v1/trail/{id}/thumbnail` (`trail_assets.is_thumbnail`) or else the first photo, not this index.
 *         like_count:
 *           type: integer
 *           default: 0
 *         category:
 *           type: string
 *         subcategory:
 *           type: string
 *         tags:
 *           type: array
 *           items:
 *             type: string
 *         gpx:
 *           type: string
 *
 *     Comment:
 *       type: object
 *       required:
 *         - id
 *         - text
 *         - author
 *         - trail
 *       properties:
 *         id:
 *           type: string
 *           description: Comment ID (15 chars)
 *         text:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     CommentInput:
 *       type: object
 *       required:
 *         - text
 *         - author
 *         - trail
 *       properties:
 *         text:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *
 *     CommentUpdateInput:
 *       type: object
 *       properties:
 *         text:
 *           type: string
 *
 *     SummitLog:
 *       type: object
 *       required:
 *         - id
 *         - date
 *         - author
 *       properties:
 *         id:
 *           type: string
 *           description: Summit log ID (15 chars)
 *         date:
 *           type: string
 *           format: date
 *         text:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         trail:
 *           type: string
 *           description: Associated trail ID
 *         gpx:
 *           type: string
 *         distance:
 *           type: number
 *         elevation_gain:
 *           type: number
 *         elevation_loss:
 *           type: number
 *         duration:
 *           type: number
 *         photos:
 *           type: array
 *           readOnly: true
 *           description: >
 *             Derived from the photo assets linked to the summit log and not writable. Each entry is a file URL,
 *             `/api/v1/files/{assetsCollectionId}/{assetId}/{file}` for a stored photo or
 *             `/api/v1/assets/{assetId}/file` for a photo kept in a connected photo library.
 *             Filled only where the summit log comes with its photo links: embedded in a trail, for example from
 *             `GET /api/v1/trail/{id}` (`expand.summit_logs_via_trail`), in the feeds, in `GET /api/v1/summit-log`
 *             with `expand=summit_log_assets_via_summit_log.asset` (empty without it) and in
 *             `GET /api/v1/profile/{handle}/stats` for `source: summit_log`. The feed and stats of a profile on
 *             another instance give absolute URLs. The endpoints that create, read or update a single summit log
 *             leave it out.
 *             Manage photos through `PUT /api/v1/assets` and `DELETE /api/v1/assets/{id}`.
 *           items:
 *             type: string
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     StatisticActivity:
 *       allOf:
 *         - $ref: '#/components/schemas/SummitLog'
 *         - type: object
 *           required:
 *             - source
 *             - collectionId
 *             - collectionName
 *           properties:
 *             source:
 *               type: string
 *               enum: [summit_log, completed_trail]
 *             collectionId:
 *               type: string
 *             collectionName:
 *               type: string
 *
 *     SummitLogInput:
 *       type: object
 *       required:
 *         - date
 *         - author
 *       properties:
 *         id:
 *           type: string
 *           description: Optional custom ID (15 chars)
 *         date:
 *           type: string
 *           format: date
 *         text:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         gpx:
 *           type: string
 *         distance:
 *           type: number
 *         elevation_gain:
 *           type: number
 *         elevation_loss:
 *           type: number
 *         duration:
 *           type: number
 *
 *     SummitLogUpdateInput:
 *       type: object
 *       properties:
 *         date:
 *           type: string
 *           format: date
 *         text:
 *           type: string
 *         gpx:
 *           type: string
 *         distance:
 *           type: number
 *         elevation_gain:
 *           type: number
 *         elevation_loss:
 *           type: number
 *         duration:
 *           type: number
 *
 *     Waypoint:
 *       type: object
 *       required:
 *         - id
 *         - lat
 *         - lon
 *         - author
 *       properties:
 *         id:
 *           type: string
 *           description: Waypoint ID (15 chars)
 *         name:
 *           type: string
 *         description:
 *           type: string
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance_from_start:
 *           type: number
 *           description: Distance from trail start in meters
 *         icon:
 *           type: string
 *           description: Icon identifier
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         photos:
 *           type: array
 *           readOnly: true
 *           description: >
 *             Derived from the photo assets linked to the waypoint and not writable. Each entry is a file URL,
 *             `/api/v1/files/{assetsCollectionId}/{assetId}/{file}` for a stored photo or
 *             `/api/v1/assets/{assetId}/file` for a photo kept in a connected photo library.
 *             Filled only where the waypoint comes with its photo links, for example embedded in a trail from
 *             `GET /api/v1/trail/{id}` or `GET /api/v1/trail` (`expand.waypoints_via_trail`). The waypoint endpoints
 *             leave it out.
 *             Manage photos through `PUT /api/v1/assets` and `DELETE /api/v1/assets/{id}`.
 *           items:
 *             type: string
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     WaypointInput:
 *       type: object
 *       required:
 *         - lat
 *         - lon
 *         - author
 *       properties:
 *         id:
 *           type: string
 *           description: Optional custom ID (15 chars)
 *         name:
 *           type: string
 *         description:
 *           type: string
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance_from_start:
 *           type: number
 *         icon:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID (15 chars)
 *         trail:
 *           type: string
 *
 *     WaypointUpdateInput:
 *       type: object
 *       properties:
 *         name:
 *           type: string
 *         description:
 *           type: string
 *         lat:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         lon:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         distance_from_start:
 *           type: number
 *         icon:
 *           type: string
 *
 *     List:
 *       type: object
 *       required:
 *         - id
 *         - name
 *         - author
 *         - public
 *       properties:
 *         id:
 *           type: string
 *           description: List ID (15 chars)
 *         name:
 *           type: string
 *         public:
 *           type: boolean
 *         description:
 *           type: string
 *         author:
 *           type: string
 *           description: Author user ID
 *         trails:
 *           type: array
 *           items:
 *             type: string
 *             description: Trail IDs (15 chars)
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     ListInput:
 *       type: object
 *       required:
 *         - name
 *         - public
 *         - trails
 *         - author
 *       properties:
 *         id:
 *           type: string
 *           description: Optional custom ID (15 chars)
 *         name:
 *           type: string
 *           minLength: 1
 *         public:
 *           type: boolean
 *         description:
 *           type: string
 *         trails:
 *           type: array
 *           items:
 *             type: string
 *             description: Trail ID (15 chars)
 *         author:
 *           type: string
 *
 *     ListUpdateInput:
 *       type: object
 *       properties:
 *         name:
 *           type: string
 *         public:
 *           type: boolean
 *         description:
 *           type: string
 *         trails:
 *           type: array
 *           items:
 *             type: string
 *         'trails-':
 *           type: string
 *           description: Remove trail by ID
 *         'trails+':
 *           type: string
 *           description: Add trail by ID
 *
 *     TrailShare:
 *       type: object
 *       required:
 *         - id
 *         - actor
 *         - trail
 *         - permission
 *       properties:
 *         id:
 *           type: string
 *           description: Share ID (15 chars)
 *         actor:
 *           type: string
 *           format: uri
 *           description: ActivityPub actor IRI
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     TrailShareInput:
 *       type: object
 *       required:
 *         - actor
 *         - trail
 *         - permission
 *       properties:
 *         actor:
 *           type: string
 *           format: uri
 *           description: ActivityPub actor IRI
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     TrailShareUpdateInput:
 *       type: object
 *       required:
 *         - permission
 *       properties:
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     ListShare:
 *       type: object
 *       required:
 *         - id
 *         - actor
 *         - list
 *         - permission
 *       properties:
 *         id:
 *           type: string
 *           description: Share ID (15 chars)
 *         actor:
 *           type: string
 *           format: uri
 *           description: ActivityPub actor IRI
 *         list:
 *           type: string
 *           description: List ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     ListShareInput:
 *       type: object
 *       required:
 *         - actor
 *         - list
 *         - permission
 *       properties:
 *         actor:
 *           type: string
 *           format: uri
 *           description: ActivityPub actor IRI
 *         list:
 *           type: string
 *           description: List ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     ListShareUpdateInput:
 *       type: object
 *       required:
 *         - permission
 *       properties:
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     TrailLinkShare:
 *       type: object
 *       required:
 *         - id
 *         - trail
 *         - permission
 *       properties:
 *         id:
 *           type: string
 *           description: Share ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     TrailLinkShareInput:
 *       type: object
 *       required:
 *         - trail
 *         - permission
 *       properties:
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     TrailLinkShareUpdateInput:
 *       type: object
 *       required:
 *         - permission
 *       properties:
 *         permission:
 *           type: string
 *           enum: [view, edit]
 *
 *     TrailLike:
 *       type: object
 *       required:
 *         - id
 *         - actor
 *         - trail
 *       properties:
 *         id:
 *           type: string
 *           description: Like ID (15 chars)
 *         actor:
 *           type: string
 *           description: User ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *         created:
 *           type: string
 *           format: date-time
 *
 *     TrailLikeInput:
 *       type: object
 *       required:
 *         - actor
 *         - trail
 *       properties:
 *         actor:
 *           type: string
 *           description: User ID (15 chars)
 *         trail:
 *           type: string
 *           description: Trail ID (15 chars)
 *
 *     Follow:
 *       type: object
 *       required:
 *         - id
 *         - follower
 *         - followee
 *       properties:
 *         id:
 *           type: string
 *           description: Follow ID (15 chars)
 *         follower:
 *           type: string
 *           description: Follower user ID
 *         followee:
 *           type: string
 *           description: Followee user ID (15 chars)
 *         created:
 *           type: string
 *           format: date-time
 *
 *     FollowInput:
 *       type: object
 *       required:
 *         - followee
 *       properties:
 *         followee:
 *           type: string
 *           description: Followee user ID (15 chars)
 *
 *     Notification:
 *       type: object
 *       required:
 *         - id
 *         - user
 *         - type
 *         - seen
 *       properties:
 *         id:
 *           type: string
 *           description: Notification ID (15 chars)
 *         user:
 *           type: string
 *           description: User ID (15 chars)
 *         type:
 *           type: string
 *           description: Notification type
 *         seen:
 *           type: boolean
 *           default: false
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     NotificationUpdateInput:
 *       type: object
 *       required:
 *         - seen
 *       properties:
 *         seen:
 *           type: boolean
 *           const: true
 *           description: Mark as read
 *
 *     Settings:
 *       type: object
 *       required:
 *         - id
 *       properties:
 *         id:
 *           type: string
 *           description: Settings ID (15 chars)
 *         unit:
 *           type: string
 *           enum: [metric, imperial]
 *         language:
 *           type: string
 *           description: Language code
 *         bio:
 *           type: string
 *           nullable: true
 *         mapFocus:
 *           type: string
 *           enum: [trails, location]
 *         location:
 *           type: object
 *           properties:
 *             name:
 *               type: string
 *             lat:
 *               type: number
 *             lon:
 *               type: number
 *           nullable: true
 *         category:
 *           type: string
 *         tilesets:
 *           type: array
 *           items:
 *             type: object
 *             properties:
 *               name:
 *                 type: string
 *               url:
 *                 type: string
 *                 format: uri
 *           nullable: true
 *         terrain:
 *           type: object
 *           properties:
 *             terrain:
 *               type: string
 *               format: uri
 *             hillshading:
 *               type: string
 *               format: uri
 *           nullable: true
 *         user:
 *           type: string
 *         privacy:
 *           type: object
 *           properties:
 *             account:
 *               type: string
 *               enum: [public, private]
 *             trails:
 *               type: string
 *               enum: [public, private]
 *             lists:
 *               type: string
 *               enum: [public, private]
 *           nullable: true
 *         created:
 *           type: string
 *           format: date-time
 *         updated:
 *           type: string
 *           format: date-time
 *
 *     SettingsInput:
 *       type: object
 *       properties:
 *         unit:
 *           type: string
 *           enum: [metric, imperial]
 *         language:
 *           type: string
 *         bio:
 *           type: string
 *           nullable: true
 *         mapFocus:
 *           type: string
 *           enum: [trails, location]
 *         location:
 *           type: object
 *           properties:
 *             name:
 *               type: string
 *             lat:
 *               type: number
 *             lon:
 *               type: number
 *           nullable: true
 *         category:
 *           type: string
 *         tilesets:
 *           type: array
 *           items:
 *             type: object
 *             properties:
 *               name:
 *                 type: string
 *               url:
 *                 type: string
 *                 format: uri
 *           nullable: true
 *         terrain:
 *           type: object
 *           properties:
 *             terrain:
 *               type: string
 *               format: uri
 *             hillshading:
 *               type: string
 *               format: uri
 *           nullable: true
 *         user:
 *           type: string
 *         privacy:
 *           type: object
 *           properties:
 *             account:
 *               type: string
 *               enum: [public, private]
 *             trails:
 *               type: string
 *               enum: [public, private]
 *             lists:
 *               type: string
 *               enum: [public, private]
 *           nullable: true
 *
 *     TrailMergeSettings:
 *       type: object
 *       required:
 *         - summitLog
 *         - photos
 *         - comments
 *         - delete
 *         - tags
 *         - likes
 *       properties:
 *         summitLog:
 *           type: boolean
 *         photos:
 *           type: boolean
 *         comments:
 *           type: boolean
 *         delete:
 *           type: boolean
 *         tags:
 *           type: boolean
 *         likes:
 *           type: boolean
 *
 *     TrailMergeSuggestRequest:
 *       type: object
 *       required:
 *         - mode
 *       properties:
 *         mode:
 *           type: string
 *           enum: [manual-selection, auto-discovery, maintenance-groups]
 *         trailIds:
 *           type: array
 *           items:
 *             type: string
 *           description: Required for manual-selection
 *         sourceTrailId:
 *           type: string
 *           description: Required for auto-discovery
 *
 *     TrailMergeSuggestCandidate:
 *       type: object
 *       required:
 *         - trailId
 *         - score
 *         - reason
 *         - warnings
 *         - selectable
 *       properties:
 *         trailId:
 *           type: string
 *         score:
 *           type: number
 *         reason:
 *           type: string
 *         warnings:
 *           type: array
 *           items:
 *             type: string
 *         selectable:
 *           type: boolean
 *
 *     TrailMergeSuggestResponse:
 *       type: object
 *       required:
 *         - targetTrailId
 *         - reason
 *         - warnings
 *         - candidates
 *       properties:
 *         targetTrailId:
 *           type: string
 *         reason:
 *           type: string
 *         warnings:
 *           type: array
 *           items:
 *             type: string
 *         candidates:
 *           type: array
 *           items:
 *             $ref: '#/components/schemas/TrailMergeSuggestCandidate'
 *
 *     TrailMergeSuggestGroup:
 *       type: object
 *       required:
 *         - groupId
 *         - trailIds
 *         - targetTrailId
 *         - reason
 *         - score
 *         - indirect
 *       properties:
 *         groupId:
 *           type: string
 *         trailIds:
 *           type: array
 *           items:
 *             type: string
 *         targetTrailId:
 *           type: string
 *         reason:
 *           type: string
 *         score:
 *           type: number
 *         indirect:
 *           type: boolean
 *
 *     TrailMergeSuggestGroupsResponse:
 *       type: object
 *       required:
 *         - groups
 *       properties:
 *         groups:
 *           type: array
 *           items:
 *             $ref: '#/components/schemas/TrailMergeSuggestGroup'
 *
 *     TrailMergeExecuteRequest:
 *       type: object
 *       required:
 *         - sourceTrailId
 *         - targetTrailId
 *         - settings
 *       properties:
 *         sourceTrailId:
 *           type: string
 *         targetTrailId:
 *           type: string
 *         settings:
 *           $ref: '#/components/schemas/TrailMergeSettings'
 *
 *     TrailMergeExecuteResponse:
 *       type: object
 *       required:
 *         - acknowledged
 *       properties:
 *         acknowledged:
 *           type: boolean
 *
 *     ListResult:
 *       type: object
 *       properties:
 *         page:
 *           type: integer
 *         perPage:
 *           type: integer
 *         totalItems:
 *           type: integer
 *         totalPages:
 *           type: integer
 *         items:
 *           type: array
 *           items:
 *             type: object
 *
 *     Asset:
 *       type: object
 *       required:
 *         - id
 *         - type
 *         - storage_mode
 *         - author
 *       properties:
 *         id:
 *           type: string
 *         collectionId:
 *           type: string
 *         collectionName:
 *           type: string
 *         type:
 *           type: string
 *           enum: [photo]
 *         file:
 *           type: string
 *         storage_mode:
 *           type: string
 *           enum: [copy, link_private]
 *         remote_status:
 *           type: string
 *           enum: [available, missing, inaccessible]
 *         remote_checked_at:
 *           type: string
 *           format: date-time
 *         remote_missing_since:
 *           type: string
 *           format: date-time
 *         remote_error:
 *           type: string
 *         author:
 *           type: string
 *         external_provider:
 *           type: string
 *         external_id:
 *           type: string
 *         taken_at:
 *           type: string
 *           format: date-time
 *         lat:
 *           type: number
 *         lon:
 *           type: number
 *         metadata:
 *           type: object
 *
 *     AssetPluginCandidate:
 *       type: object
 *       description: >
 *         A photo offered for attaching to a trail, waypoint or summit log, either from the user's own photos in
 *         wanderer (`POST /api/v1/assets/library`) or from a photo library plugin
 *         (`POST /api/v1/plugins/assets/{plugin}/candidates`).
 *       required:
 *         - assetId
 *         - originalFileName
 *         - takenAt
 *         - lat
 *         - lon
 *         - distance
 *       properties:
 *         assetId:
 *           type: string
 *           description: The asset record ID for a photo in wanderer, otherwise the plugin's ID for the photo.
 *         originalFileName:
 *           type: string
 *         takenAt:
 *           type: string
 *           format: date-time
 *           description: The photo's time for display and sorting. It can be a file or upload time.
 *         capturedAt:
 *           type: string
 *           format: date-time
 *           description: >
 *             When the photo was taken according to its own metadata, such as its EXIF capture date. Only set when
 *             that metadata provides it, unlike `takenAt`, which can be a file or upload time. A local time without a
 *             time zone counts as UTC.
 *         lat:
 *           type: number
 *         lon:
 *           type: number
 *         distance:
 *           type: number
 *         pointLat:
 *           type: number
 *         pointLon:
 *           type: number
 *         distanceFromStart:
 *           type: number
 *         city:
 *           type: string
 *         country:
 *           type: string
 *         source:
 *           type: string
 *           description: Set to `wanderer` for a photo in wanderer.
 *         providerId:
 *           type: string
 *           description: For a photo in wanderer, `externalProvider` when set, otherwise `wanderer`.
 *         externalProvider:
 *           type: string
 *           description: For a photo in wanderer, the plugin it was imported or linked from, such as `immich`.
 *         externalId:
 *           type: string
 *           description: For a photo in wanderer, its ID at `externalProvider`.
 *         thumbnailUrl:
 *           type: string
 *           description: >
 *             For a photo in wanderer, the URL of its full-size file; add a `thumb` query such as `?thumb=300x300`
 *             for a smaller image. Plugin candidates usually leave it out; their previews come from
 *             `GET /api/v1/plugins/assets/{plugin}/thumbnail/{assetId}`.
 *
 *     PhotoLibraryBounds:
 *       type: object
 *       description: >
 *         Map area to search for photos, in degrees. `south` must not be greater than `north`; `west` greater than
 *         `east` crosses the antimeridian.
 *       required:
 *         - west
 *         - south
 *         - east
 *         - north
 *       properties:
 *         west:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         south:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *         east:
 *           type: number
 *           minimum: -180
 *           maximum: 180
 *         north:
 *           type: number
 *           minimum: -90
 *           maximum: 90
 *
 *     Error:
 *       type: object
 *       required:
 *         - message
 *       properties:
 *         message:
 *           type: string
 *         code:
 *           type: integer
 *         data:
 *           type: object
 */
export {};
