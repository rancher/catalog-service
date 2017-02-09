import pytest
import cattle
import requests
from wait_for import wait_for


def headers(environment_id):
    return {
        'Accept': 'application/json',
        'x-api-project-id': environment_id
    }


DEFAULT_HEADERS = headers('e1')


def create_catalog(name, url, branch=None, headers=DEFAULT_HEADERS):
    schemas_url = 'http://localhost:8088/v1-catalog/schemas'
    client = cattle.from_env(url=schemas_url, headers=headers)

    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    data = {
        'name': name,
        'url': url,
    }
    if branch:
        data['branch'] = branch

    api_url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.post(api_url, data=data, headers=headers)
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == name
    assert resp['url'] == url
    if branch:
        assert resp['branch'] == branch

    api_url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(api_url, headers=headers)
    assert response.status_code == 204

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs) + 1
    assert len(templates) > len(original_templates)

    return resp


def delete_catalog(name, headers=DEFAULT_HEADERS):
    schemas_url = 'http://localhost:8088/v1-catalog/schemas'
    client = cattle.from_env(url=schemas_url, headers=headers)

    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/catalogs/' + name
    response = requests.delete(url, headers=headers)
    assert response.status_code == 204

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs) - 1
    assert len(templates) < len(original_templates)


@pytest.fixture
def client():
    url = 'http://localhost:8088/v1-catalog/schemas'
    catalogs = cattle.from_env(url=url, headers=DEFAULT_HEADERS).list_catalog()
    wait_for(
        lambda: len(catalogs) > 0
    )
    return cattle.from_env(url=url, headers=DEFAULT_HEADERS)


def test_catalog_list(client):
    catalogs = client.list_catalog()
    assert len(catalogs) == 2
    assert catalogs[0].name == 'updated'
    assert catalogs[0].url == '/tmp/test-catalog'
    assert catalogs[1].name == 'orig'
    assert catalogs[1].url == 'https://github.com/rancher/test-catalog'


def test_get_catalog(client):
    url = 'http://localhost:8088/v1-catalog/catalogs/orig'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == 'orig'
    assert resp['url'] == 'https://github.com/rancher/test-catalog'


def test_catalog_commit(client):
    url = 'http://localhost:8088/v1-catalog/catalogs/orig'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['commit'] == '1af7e6801786317999f4c51a103fa65b065c7bc8'

    url = 'http://localhost:8088/v1-catalog/catalogs/updated'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['commit'] != '1af7e6801786317999f4c51a103fa65b065c7bc8'


def test_create_and_delete_catalog(client):
    url = 'https://github.com/rancher/community-catalog'
    create_catalog('created', url)
    delete_catalog('created')


def test_catalog_branch(client):
    url = 'https://github.com/rancher/test-catalog'
    create_catalog('branch', url, "test-branch")
    delete_catalog('branch')


def test_broken_catalog(client):
    url = 'https://github.com/rancher/test-catalog'
    broken_catalog = create_catalog('broken', url, "broken")
    assert broken_catalog['transitioningMessage'] != ''
    delete_catalog('broken')


def test_catalog_different_environment(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'https://github.com/rancher/community-catalog'
    create_catalog('env', url, headers=headers('e2'))

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs)
    assert len(templates) == len(original_templates)

    delete_catalog('env', headers=headers('e2'))


def test_template_list(client):
    templates = client.list_template()
    assert len(templates) > 0


def test_get_template(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['folderName'] == 'k8s'


def test_template_version_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 14

    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions' + \
        '?rancherVersion=v1.0.1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 9


def test_upgrade_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/' + \
        'orig:test-upgrade-links:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 10

    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions:2' + \
        '?rancherVersion=v1.0.1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 7


def test_template_icon(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:nfs-server?image'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    assert len(response.content) == 1139


def test_get_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['revision'] == 0

    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['revision'] == 1


def test_template_version_questions(client):
    url = 'http://localhost:8088/v1-catalog/templates/' + \
        'orig:all-question-types:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    questions = resp['questions']
    assert questions is not None
    assert len(questions) == 11

    assert questions[0]['variable'] == 'TEST_STRING'
    assert questions[0]['label'] == 'String'
    assert not questions[0]['required']
    assert questions[0]['default'] == 'hello'
    assert questions[0]['type'] == 'string'

    assert questions[1]['variable'] == 'TEST_MULTILINE'
    assert questions[1]['label'] == 'Multi-Line'
    assert not questions[1]['required']
    assert questions[1]['default'] == 'Hello\nWorld\n'
    assert questions[1]['type'] == 'multiline'

    assert questions[2]['variable'] == 'TEST_PASSWORD'
    assert questions[2]['label'] == 'Password'
    assert not questions[2]['required']
    assert questions[2]['default'] == "not-so-secret stuff"
    assert questions[2]['type'] == 'password'

    assert questions[3]['variable'] == 'TEST_ENUM'
    assert questions[3]['label'] == 'Enum'
    assert not questions[3]['required']
    assert questions[3]['options'] == ['purple', 'monkey', 'dishwasher']
    assert questions[3]['default'] == 'monkey'
    assert questions[3]['type'] == 'enum'

    assert questions[4]['variable'] == 'TEST_DATE'
    assert questions[4]['label'] == 'Date'
    assert not questions[4]['required']
    assert questions[4]['default'] == '2015-07-25T19:55:00Z'
    assert questions[4]['type'] == 'date'

    assert questions[5]['variable'] == 'TEST_INT'
    assert questions[5]['label'] == 'Integer'
    assert not questions[5]['required']
    assert questions[5]['default'] == '42'
    assert questions[5]['type'] == 'int'

    assert questions[6]['variable'] == 'TEST_FLOAT'
    assert questions[6]['label'] == 'Float'
    assert not questions[6]['required']
    assert questions[6]['default'] == '4.2'
    assert questions[6]['type'] == 'float'

    assert questions[7]['variable'] == 'TEST_BOOLEAN'
    assert questions[7]['label'] == 'Boolean'
    assert not questions[7]['required']
    assert questions[7]['default'] == 'true'
    assert questions[7]['type'] == 'boolean'

    assert questions[8]['variable'] == 'TEST_SERVICE'
    assert questions[8]['label'] == 'Service'
    assert not questions[8]['required']
    assert questions[8]['default'] == 'kopf'
    assert questions[8]['type'] == 'service'

    assert questions[9]['variable'] == 'TEST_CERTIFICATE'
    assert questions[9]['label'] == 'Certificate'
    assert not questions[9]['required']
    assert questions[9]['default'] == 'rancher.rocks'
    assert questions[9]['type'] == 'certificate'

    assert questions[10]['variable'] == 'TEST_UNKNOWN'
    assert questions[10]['label'] == 'Unknown'
    assert not questions[10]['required']
    assert questions[10]['default'] == 'wha?'
    assert questions[10]['type'] == 'unknown'


def test_template_version_bindings(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    bindings = resp['bindings']
    assert bindings is not None


def test_refresh(client):
    url = 'http://localhost:8088/v1-catalog/templates/updated:many-versions:14'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == '1.0.14'


def test_refresh_no_changes(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 204

    catalogs = client.list_catalog()
    templates = client.list_template()
    assert len(catalogs) == len(original_catalogs)
    assert len(templates) == len(original_templates)


def test_v2_syntax(client):
    for revision in [0, 1, 2, 3]:
        url = 'http://localhost:8088/v1-catalog/templates/orig:v2:' + \
            str(revision)
        response = requests.get(url, headers=DEFAULT_HEADERS)
        assert response.status_code == 200
