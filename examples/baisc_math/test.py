class SimpleApp:
    def __init__(self):
        self.routes = {}

    def route(self, path):
        def decorator(func):
            self.routes[path] = func
            return func
        return decorator

    def execute(self, path):
        if path in self.routes:
            return self.routes[path]()
        else:
            raise ValueError("Route not found!")

app = SimpleApp()

@app.route("/")
def home():
    return "Welcome to the home page!"

@app.route("/about")
def about():
    return "About us page!"
