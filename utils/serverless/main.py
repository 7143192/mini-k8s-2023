import os
import json
import function

from flask import Flask, request

app = Flask(__name__)


@app.route("/", methods=['POST'])
def work():
    try:
        params = json.loads(request.get_data())
    except json.JSONDecodeError:
        params = ""
    finally:
        res = function.main(params)
        return json.dumps(res)


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=int(os.environ.get('PORT', 9090)))