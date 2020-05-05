from behave import step
import os
import json
import requests
from hamcrest import assert_that, equal_to, is_not
from warrant import Cognito

config = dict(
    domain = os.environ["JOURNAL_BASE_DOMAIN"],
)

@step('we issue a "{method}" to "{url}"')
@step('we issue a "{method}" to "{url}" with payload')
def rest_call(context, method, url, unwrapped=None):
    payload = context.text
    if payload:
        payload = payload%{**config, **context.fields}
        json.dumps(json.loads(context.text))

    url = url%{**config, **context.fields}
    context.req = {
        "method": method,
        "url": url,
        "payload": payload,
    }
    context.resp = requests.request(
        method,
        url,
        data=payload,
        verify=False,
    )

@step('the response http code is {code:d}')
def check_http_code(context, code):
    assert_that(context.resp.status_code, equal_to(code))

@step('the response payload resembles')
def check_response_payload(context):
    text = context.text%{**config, **context.fields}
    want = json.loads(text)
    have = context.resp.json()
    context.resembles = True

    if isinstance(have, dict):
        check_dict_r(context, want, have)
    elif isinstance(have, list):
        check_list_r(context, "top-level", want, have)
    else:
        assert False, "called this step incorrectly"

@step('the response payload does not resemble')
def check_response_payload_isnt(context):
    text = context.text%{**config, **context.fields}
    want = json.loads(text)
    have = context.resp.json()
    context.resembles = False

    if isinstance(have, dict):
        check_dict_r(context, want, have)
    elif isinstance(have, list):
        check_list_r(context, "top-level", want, have)
    else:
        assert False, "called this step incorrectly"

@step('we store the response field {field} as {key}')
def store_response_field_as(context, field, key):
    context.fields[key] = context.resp.json()[field]

@step('we store the response field {field}')
def store_response_field(context, field):
    context.fields[field] = context.resp.json()[field]

@step('this query returns {want_rows:d} row')
@step('this query returns {want_rows:d} rows')
def query_returns(context, want_rows):
    c = context.db.cursor()
    query = context.text%{**config, **context.fields}
    c.execute(query)
    have_rows = c.rowcount
    assert_that(have_rows, equal_to(want_rows))

@step('we execute this query')
def execute_query(context):
    c = context.db.cursor()
    c.execute(context.text)

def check_dict_r(context, want, have):
    for k in want:
        assert k in have, "expected key:%s to exist in %s"%(k, have.keys())
        check_type_r(context, k, want[k], have[k])

def check_list_r(context, key, want, have):
    for w in want:
        found = False
        for h in have:
            try:
                check_type_r(context, key, w, h)
                found = True
                break
            except AssertionError:
                continue
        if context.resembles:
            assert found, "expected %s list to have element %s, but it didn't"%(key, w)
        else:
            assert not found, "expected %s list NOT to have element %s, but it DID"%(key, w)

def check_type_r(context, key, want, have):
    if isinstance(want, dict):
        check_dict_r(context, want, have)
    elif isinstance(want, list):
        check_list_r(context, key, want, have)
    else:
        if context.resembles:
            assert_that(have, equal_to(want))
        else:
            assert_that(have, is_not(equal_to(want)))
