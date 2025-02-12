import json
from http.server import HTTPServer, BaseHTTPRequestHandler

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
    server.serve_forever()

if __name__ == "__main__":
    main()
