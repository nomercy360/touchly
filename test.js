const {spec} = require('pactum');
const {faker} = require('@faker-js/faker');

const API_URL = 'http://127.0.0.1:8080/api';
const ADMIN_URL = 'http://127.0.0.1:8080/admin';

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

    const contact = {
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

    it('POST /contacts', async () => {
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
                tags: [
                    {id: '$S{firstTagId}'},
                    {id: '$S{secondTagId}'}
                ],
                social_links: [
                    {type: 'linkedin', link: contact.social_links[0].link},
                    {type: 'twitter', link: contact.social_links[1].link}
                ]
            })
            .stores('contactId', 'id');
    });

    const address = {
        external_id: faker.string.uuid(),
        label: faker.word.noun(),
        name: faker.location.street(),
        location: {
            lat: faker.location.latitude(),
            lng: faker.location.longitude()
        },
        contact_id: '$S{contactId}'
    }

    it('POST /contacts/:contactId/address', async () => {
        await spec()
            .post(API_URL + '/contacts/' + '$S{contactId}' + '/address')
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
                contact_id: '$S{contactId}'
            })
            .stores('addressId', 'id');
    });

    const search = '?lat=37.7749&lng=-122.4194&radius=10000';

    it('GET /contacts', async () => {
        await spec()
            .get(API_URL + '/contacts' + search)
            .expectStatus(200)
            .expectJsonMatch({
                page: 1,
                page_size: 20,
                total_count: 1,
                contacts: [
                    {
                        id: '$S{contactId}',
                        name: contact.name,
                        avatar: contact.avatar,
                        activity_name: contact.activity_name,
                        about: contact.about,
                        views_amount: 0,
                        saves_amount: 0,
                        user_id: '$S{userId}',
                        is_published: false,
                    }
                ]
            });
    });
});