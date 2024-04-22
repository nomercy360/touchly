const {spec} = require('pactum');
const {faker} = require('@faker-js/faker');

const API_URL = 'http://127.0.0.1:8080/api';
const ADMIN_URL = 'http://127.0.0.1:8080/admin';
const CDN_URL = 'https://assets.mxksim.dev';

const TEST_USER = {
    email: faker.internet.email(),
    password: faker.internet.password()
}

describe('API Test', () => {
    before(async () => {
        await spec()
            .post(ADMIN_URL + '/users')
            .withJson({
                email: TEST_USER.email,
                password: TEST_USER.password
            })
            .withHeaders({
                'Content-Type': 'application/json',
                'X-Api-Key': 'secret'
            })
            .expectStatus(201)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'email', 'created_at', 'updated_at', 'email_verified_at', 'deleted_at']
            })
            .expectJsonMatch({
                email: TEST_USER.email,
                deleted_at: null,
            })
            .stores('userId', 'id');

        await spec()
            .post(API_URL + '/login')
            .withJson({
                email: TEST_USER.email,
                password: TEST_USER.password

            })
            .expectStatus(200)
            .expectJsonSchema({
                type: 'object',
                required: ['token']
            })
            .stores('token', 'token');
    });

    it('GET /me', async () => {
        await spec()
            .get(API_URL + '/me')
            .withBearerToken('$S{token}')
            .expectStatus(200)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'email', 'created_at', 'updated_at', 'email_verified_at', 'deleted_at']
            })
            .expectJsonMatch({
                email: TEST_USER.email,
                deleted_at: null,
            });
    });

    it('POST /tags', async () => {
        const firstTag = faker.word.noun()

        await spec()
            .post(API_URL + '/tags')
            .withJson({name: firstTag})
            .withBearerToken('$S{token}')
            .expectStatus(201)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'name']
            })
            .expectJsonMatch({
                name: firstTag
            })
            .stores('firstTagId', 'id');

        const secondTag = faker.word.noun()

        await spec()
            .post(API_URL + '/tags')
            .withJson({name: secondTag})
            .withBearerToken('$S{token}')
            .expectStatus(201)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'name']
            })
            .expectJsonMatch({
                name: secondTag
            })
            .stores('secondTagId', 'id');
    });

    it('GET /tags', async () => {
        await spec()
            .get(API_URL + '/tags')
            .expectStatus(200)
            .expectJsonLength(2)
            .expectJsonSchema({
                type: 'array',
                items: {
                    type: 'object',
                    required: ['id', 'name']
                }
            });
    });

    const firstContact = {
        name: faker.person.fullName(),
        avatar: faker.image.avatar(),
        activity_name: faker.company.name(),
        about: faker.lorem.paragraph(),
        website: faker.internet.url(),
        country_code: faker.location.countryCode(),
        phone_number: faker.phone.number(),
        phone_calling_code: '+1',
        email: faker.internet.email(),
        tags: [
            {id: '$S{firstTagId}'},
            {id: '$S{secondTagId}'}
        ],
        social_links: [
            {type: 'linkedin', link: faker.internet.url()},
            {type: 'twitter', link: faker.internet.url()}
        ]
    }

    const secondContact = {
        name: faker.person.fullName(),
        avatar: faker.image.avatar(),
        activity_name: faker.company.name(),
        about: faker.lorem.paragraph(),
        website: faker.internet.url(),
        country_code: faker.location.countryCode(),
        phone_number: faker.phone.number(),
        phone_calling_code: '+1',
        email: faker.internet.email(),
        tags: [
            {id: '$S{secondTagId}'}
        ],
        social_links: [
            {type: 'linkedin', link: faker.internet.url()},
            {type: 'twitter', link: faker.internet.url()}
        ]
    }

    it('POST /contacts', async () => {
        for (let contact of [firstContact, secondContact]) {
            let storeVarName = contact === firstContact ? 'firstContactId' : 'secondContactId';

            await spec()
                .post(API_URL + '/contacts')
                .withJson(contact)
                .withBearerToken('$S{token}')
                .expectStatus(201)
                .expectJsonSchema({
                    type: 'object',
                    required: ['id', 'name', 'avatar', 'activity_name', 'about', 'website', 'country_code', 'phone_number', 'phone_calling_code', 'email', 'user_id', 'tags', 'social_links']
                })
                .expectJsonMatch({
                    name: contact.name,
                    avatar: contact.avatar,
                    activity_name: contact.activity_name,
                    about: contact.about,
                    website: contact.website,
                    country_code: contact.country_code,
                    phone_number: contact.phone_number,
                    phone_calling_code: contact.phone_calling_code,
                    email: contact.email,
                    tags: contact.tags,
                    social_links: contact.social_links,
                })
                .stores(storeVarName, 'id');
        }
    });

    const firstContactAddress = {
        external_id: faker.string.uuid(),
        label: faker.word.noun(),
        name: faker.location.street(),
        location: {
            // Somewhere in Moscow
            lat: faker.location.latitude({min: 55.7558, max: 55.8558}),
            lng: faker.location.longitude({min: 37.6176, max: 37.7176})
        },
        contact_id: '$S{firstContactId}'
    }

    const secondContactAddress = {
        external_id: faker.string.uuid(),
        label: faker.word.noun(),
        name: faker.location.street(),
        location: {
            // Somewhere in London
            lat: faker.location.latitude({min: 51.5074, max: 51.6074}),
            lng: faker.location.longitude({min: -0.2278, max: -0.1278})
        },
        contact_id: '$S{secondContactId}'
    }

    it('POST /contacts/:contactId/address', async () => {
        for (let address of [firstContactAddress, secondContactAddress]) {
            let storeVarName = address === firstContactAddress ? 'firstAddressId' : 'secondAddressId';
            let contactId = address === firstContactAddress ? '$S{firstContactId}' : '$S{secondContactId}';

            await spec()
                .post(API_URL + '/contacts/' + contactId + '/address')
                .withJson(address)
                .withBearerToken('$S{token}')
                .expectStatus(201)
                .expectJsonSchema({
                    type: 'object',
                    required: ['id', 'external_id', 'label', 'name', 'location', 'contact_id']
                })
                .expectJsonMatch({
                    external_id: address.external_id,
                    label: address.label,
                    name: address.name,
                    location: address.location,
                    contact_id: contactId
                })
                .stores(storeVarName, 'id');
        }
    });

    it('GET /contacts', async () => {
        await spec()
            .get(API_URL + '/contacts')
            .expectStatus(200)
            .expectJsonSchema({
                type: 'object',
                required: ['page', 'page_size', 'total_count', 'contacts']
            })
            .expectJsonMatch({
                page: 1,
                page_size: 20,
                total_count: 2
            })
            .expectJsonLength('contacts', 2);
    });

    const searchCases = {
        'Search By Name': {
            query: '?search=' + firstContact.name,
            expected: {
                page: 1,
                page_size: 20,
                total_count: 1,
                contacts: [
                    {
                        id: '$S{firstContactId}',
                        name: firstContact.name,
                        avatar: firstContact.avatar,
                        activity_name: firstContact.activity_name,
                        about: firstContact.about,
                        views_amount: 0,
                        saves_amount: 0,
                        user_id: '$S{userId}',
                        is_published: false,
                    }
                ]
            },
        },
        'Search By Location': {
            query: '?lat=55.7558&lng=37.6176&radius=100',
            expected: {
                page: 1,
                page_size: 20,
                total_count: 1,
                contacts: [
                    {
                        id: '$S{firstContactId}',
                        name: firstContact.name,
                        avatar: firstContact.avatar,
                        activity_name: firstContact.activity_name,
                        about: firstContact.about,
                        views_amount: 0,
                        saves_amount: 0,
                        user_id: '$S{userId}',
                        is_published: false,
                    }
                ]
            },
        },
        'Search By Tag': {
            query: '?tag=$S{firstTagId}',
            expected: {
                page: 1,
                page_size: 20,
                total_count: 1,
                contacts: [
                    {
                        id: '$S{firstContactId}',
                        name: firstContact.name,
                        avatar: firstContact.avatar,
                        activity_name: firstContact.activity_name,
                        about: firstContact.about,
                        views_amount: 0,
                        saves_amount: 0,
                        user_id: '$S{userId}',
                        is_published: false,
                    }
                ]
            },
        },
    }

    for (let [name, data] of Object.entries(searchCases)) {
        it(name + ' GET /contacts' + data.query, async () => {
            await spec()
                .get(API_URL + '/contacts' + data.query)
                .expectStatus(200)
                .expectJsonSchema({
                    type: 'object',
                    required: ['page', 'page_size', 'total_count', 'contacts']
                })
                .expectJsonMatch(data.expected);
        });
    }

    it('GET /contacts/:contactId', async () => {
        await spec()
            .get(API_URL + '/contacts/$S{firstContactId}')
            .expectStatus(200)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'name', 'avatar', 'activity_name', 'about', 'website', 'country_code', 'phone_number', 'phone_calling_code', 'email', 'user_id', 'tags', 'social_links', 'views_amount', 'saves_amount', 'is_published']
            })
            .expectJsonMatch({
                id: '$S{firstContactId}',
                name: firstContact.name,
                avatar: firstContact.avatar,
                activity_name: firstContact.activity_name,
                about: firstContact.about,
                website: firstContact.website,
                country_code: firstContact.country_code,
                phone_number: firstContact.phone_number,
                phone_calling_code: firstContact.phone_calling_code,
                email: firstContact.email,
                tags: firstContact.tags,
                social_links: firstContact.social_links,
                views_amount: 0,
                saves_amount: 0,
                user_id: '$S{userId}',
                is_published: false,
            });
    });

    const firstContactUpdate = {
        name: faker.person.fullName(),
        avatar: faker.image.avatar(),
        activity_name: faker.company.name(),
        website: faker.internet.url(),
        tags: [
            {id: '$S{secondTagId}'}
        ],
        social_links: [
            {type: 'instagram', link: faker.internet.url()},
            {type: 'facebook', link: faker.internet.url()},
            {type: 'twitter', link: faker.internet.url()}
        ]
    }

    it('PUT /contacts/:contactId', async () => {
        await spec()
            .put(API_URL + '/contacts/$S{firstContactId}')
            .withJson(firstContactUpdate)
            .withBearerToken('$S{token}')
            .expectStatus(200)
            .expectJsonSchema({
                type: 'object',
                required: ['id', 'name', 'avatar', 'activity_name', 'about', 'website', 'country_code', 'phone_number', 'phone_calling_code', 'email', 'user_id', 'tags', 'social_links', 'views_amount', 'saves_amount', 'is_published']
            })
            .expectJsonMatch({
                name: firstContactUpdate.name,
                avatar: firstContactUpdate.avatar,
                activity_name: firstContactUpdate.activity_name,
                about: firstContact.about,
                website: firstContactUpdate.website,
                country_code: firstContact.country_code,
                phone_number: firstContact.phone_number,
                phone_calling_code: firstContact.phone_calling_code,
                email: firstContact.email,
                tags: firstContactUpdate.tags,
                social_links: firstContactUpdate.social_links,
                views_amount: 0,
                saves_amount: 0,
                user_id: '$S{userId}',
                is_published: false,
            });
    });


    it('POST /uploads/get-url', async () => {
        const fileName = faker.system.commonFileName('png');

        await spec()
            .post(API_URL + '/uploads/get-url?file_name=' + fileName)
            .withBearerToken('$S{token}')
            .expectStatus(200)
            .withRequestTimeout(5000)
            .expectJsonSchema({
                type: 'object',
                required: ['url']
            })
            .stores('url', 'url');

        await spec()
            .put('$S{url}')
            .withFile('test.png', './test-data/test-pic.png', {contentType: 'image/png'})
            .withRequestTimeout(10000)
            .expectStatus(200);

        // File should be uploaded to S3
        await spec()
            .get(CDN_URL + '/$S{userId}/' + fileName)
            .withRequestTimeout(10000)
            .expectStatus(200)
            .save('res-pic.png');
    });
});