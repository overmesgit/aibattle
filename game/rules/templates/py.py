import json
from http.server import HTTPServer, BaseHTTPRequestHandler

class Position:
    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y

    def distance_to(self, other) -> int:
        return ((other.x-self.x)**2 + (other.y-self.y)**2)**0.5

    def to_dict(self):
        return {"x": self.x, "y": self.y}

class Action:
    def __init__(self, action: str, target: Position, error: str = ""):
        self.action = action
        self.target = target
        self.error = error

    def to_dict(self):
        return {
            "action": self.action,
            "target": self.target.to_dict(),
            "error": self.error
        }

class ActionResponse:
    def __init__(self, actions: list):
        self.unit_action = actions

    def to_dict(self):
        return {
            "unit_action": [action.to_dict() for action in self.unit_action]
        }


# implementation of get_turn_actions(input_dict)
<generated>

def play_turn(input_json: bytes) -> bytes:
    input_dict = json.loads(input_json)
    response = get_turn_actions(input_dict)
    print(f"input: {input_dict}")
    print(f"response: {response}")
    return json.dumps(response.to_dict()).encode()

class GameHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)

        try:
            response = play_turn(body)
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(response)
        except Exception as e:
            self.send_response(500)
            self.send_header('Content-Type', 'text/plain')
            self.end_headers()
            self.wfile.write(str(e).encode())

def main():
    print("starting game server")
    server = HTTPServer(('', 8080), GameHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("shutting down server")
        server.server_close()

if __name__ == "__main__":
    main()
