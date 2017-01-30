import pytest
import cattle
import requests
import time
from wait_for import wait_for


@pytest.fixture
def client():
    time.sleep(8)
    url = 'http://localhost:8088/v1-catalog/schemas'
    catalogs = cattle.from_env(url=url).list_catalog()
    wait_for(
        lambda: len(catalogs) > 0
    )
    return cattle.from_env(url=url)


def test_catalog_list(client):
    catalogs = client.list_catalog()
    assert len(catalogs) > 0
    # assert catalogs[0].name == 't'
    # assert catalogs[1].name == 't2'
    # assert catalogs[2].name == 'library'


def test_get_catalog(client):
    url = 'http://localhost:8088/v1-catalog/catalogs/t'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == 't'
    assert resp['url'] == 'https://github.com/joshwget/test-catalog'


def test_create_catalog(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.post(url, data={
        'name': 'created',
        'url': 'https://github.com/rancher/community-catalog',
    })
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == 'created'
    assert resp['url'] == 'https://github.com/rancher/community-catalog'

    url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(url)
    assert response.status_code == 204

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs) + 1
    assert len(templates) > len(original_templates)


def test_template_list(client):
    templates = client.list_template()
    assert len(templates) > 0


def test_get_template(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:k8s'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert resp['folderName'] == 'k8s'


def test_template_version_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:many-versions'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 14

    url = 'http://localhost:8088/v1-catalog/templates/t:many-versions' + \
            '?rancherVersion=v1.0.1'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 9


def test_template_icon(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:nfs-server?image'
    response = requests.get(url)
    assert response.status_code == 200
    assert len(response.content) == 1139


def test_get_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:k8s:1'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert resp['revision'] == 1


def test_template_bindings(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:k8s:1'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    bindings = resp['bindings']
    assert bindings is not None


def test_refresh(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:many-versions:14'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == '1.0.14'


def test_refresh_no_changes(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(url)
    assert response.status_code == 204

    catalogs = client.list_catalog()
    templates = client.list_template()
    assert len(catalogs) == len(original_catalogs)
    assert len(templates) == len(original_templates)


def test_v2_syntax(client):
    for revision in [0, 1, 2, 3]:
        url = 'http://localhost:8088/v1-catalog/templates/t:v2:' + \
                str(revision)
        response = requests.get(url)
        assert response.status_code == 200


def test_upgrade_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/t:test-upgrade-links:1'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 11

    url = 'http://localhost:8088/v1-catalog/templates/t:many-versions:2' + \
            '?rancherVersion=v1.0.1'
    response = requests.get(url)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 7
