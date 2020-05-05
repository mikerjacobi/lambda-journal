import sys
sys.path.append("features/steps")
import common
import os

BEHAVE_DEBUG_ON_ERROR = False

def before_all(context):
    global BEHAVE_DEBUG_ON_ERROR, BEHAVE_RETAIN_DB_DATA
    BEHAVE_DEBUG_ON_ERROR = context.config.userdata.getbool("DEBUG")

def before_feature(context, step):
    context.fields = {}

def after_step(context, step):
    if BEHAVE_DEBUG_ON_ERROR and step.status == "failed":
        import ipdb, sys
        sys.stdout = sys.__stdout__
        ipdb.post_mortem(step.exc_traceback)




