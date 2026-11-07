import ast
from pathlib import Path
import commerce_http
import inference_lifecycle
import managed_inference
import gce_provider

BASE=Path(__file__).resolve().parents[2]/'base-00007'

def test_ready_window_stream_ast_is_preserved_when_ready_route_added():
    def stream_tree(path):
        tree=ast.parse(Path(path).read_text())
        return ast.dump(next(n for n in ast.walk(tree) if isinstance(n,ast.FunctionDef) and n.name=='stream'),include_attributes=False)
    assert stream_tree(commerce_http.__file__)==stream_tree(BASE/'commerce_http.py')

def test_lease_and_managed_ready_window_modules_unchanged_by_rebase():
    for module in (inference_lifecycle,managed_inference):
        assert Path(module.__file__).read_bytes()==(BASE/Path(module.__file__).name).read_bytes()

def test_discard_provider_unchanged_by_rebase():
    assert Path(gce_provider.__file__).read_bytes()==(BASE/'gce_provider.py').read_bytes()
