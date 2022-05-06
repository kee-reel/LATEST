from marshmallow import Schema, fields, validates_schema, ValidationError

from errors import ERROR

class SolutionData(Schema):
    text = fields.Str()
    extention = fields.Str()

class SolutionTests(Schema):
    fixed = fields.Str()
    random = fields.Str()

class Task(Schema):
    id = fields.Int()
    user_solution = fields.Nested(SolutionData)
    complete_solution = fields.Nested(SolutionData)
    tests = fields.Nested(SolutionTests)

class BuildErr(Schema):
    msg = fields.Str()

class TimeoutErr(Schema):
    params = fields.Str()
    time = fields.Float()

class RuntimeErr(Schema):
    params = fields.Str()
    msg = fields.Str()

class TestErr(Schema):
    params = fields.Str()
    expected = fields.Str()
    result = fields.Str()

class TestErrData(Schema):
    tests_passed = fields.Int(required=True)
    tests_total = fields.Int(required=True)
    build = fields.Nested(BuildErr)
    timeout = fields.Nested(TimeoutErr)
    runtime = fields.Nested(RuntimeErr)
    test = fields.Nested(TestErr)

    _one_of = {'build', 'timeout', 'runtime', 'test'}

    @validates_schema
    def test_error_fields(self, data, **kwargs):
        if len(set(data.keys()) & self._one_of) != 1:
            raise ValidationError(f'Malformed error data: {data}')

class Result(Schema):
    id = fields.Int()
    error_data = fields.Nested(TestErrData)

