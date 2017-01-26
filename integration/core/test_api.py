import pytest
import cattle
import requests
import time
from wait_for import wait_for


@pytest.fixture
def client():
    time.sleep(7)
    url = 'http://localhost:8088/v1-catalog/schemas'
    catalogs = cattle.from_env(url=url).list_catalog()
    wait_for(
        lambda: len(catalogs) > 0
    )
    return cattle.from_env(url=url)


def test_catalog_list(client):
    catalogs = client.list_catalog()
    assert len(catalogs) == 2
    assert catalogs[0].name == 't'
    assert catalogs[1].name == 'library'


# def test_get_catalog(client):
#     url = 'http://localhost:8088/v1-catalog/catalogs/t'
#     response = requests.get(url)
#     assert response.status_code == 200
#     resp = response.json()
#     assert resp['name'] == 't'
#     assert resp['url'] == 'https://github.com/joshwget/test-catalog'


def test_create_catalog(client):
    originalTemplates = client.list_template()
    assert len(originalTemplates) > 0

    url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.post(url, data={
        'name': 't2',
        'url': 'https://github.com/rancher/community-catalog',
    })
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == 't2'
    assert resp['url'] == 'https://github.com/rancher/community-catalog'

    url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(url)
    assert response.status_code == 204

    templates = client.list_template()
    assert len(templates) > len(originalTemplates)


# def test_template_list(client):
#     templates = client.list_template()
#     assert len(templates) > 0
#
#
# def test_template_basics(client):
#     url = 'http://localhost:8088/v1-catalog/templates/t:k8s:0'
#     response = requests.get(url)
#     assert response.status_code == 200
#     resp = response.json()
#     assert resp['template'] == 'k8s'
#
#
# def test_template_icon(client):
#     url = 'http://localhost:8088/v1-catalog/templates/t:nfs-server?image'
#     response = requests.get(url)
#     assert response.status_code == 200
#     assert len(response.content) == 1139
#
#
# def test_template_bindings(client):
#     url = 'http://localhost:8088/v1-catalog/templates/t:k8s:0'
#     response = requests.get(url)
#     assert response.status_code == 200
#     resp = response.json()
#     bindings = resp['bindings']
#     assert bindings is not None
#
#
# def test_v2_syntax(client):
#     for revision in [0, 1, 2, 3]:
#         url = 'http://localhost:8088/v1-catalog/templates/t:v2:' + \
#                 str(revision)
#         response = requests.get(url)
#         assert response.status_code == 200
#
#
# def test_upgrade_links(client):
#     url = 'http://localhost:8088/v1-catalog/templates/t:test-upgrade-links:1'
#     response = requests.get(url)
#     assert response.status_code == 200
#     resp = response.json()
#     print resp
#     upgradeLinks = resp['upgradeVersionLinks']
#     assert upgradeLinks is not None
#     assert len(upgradeLinks) == 11
