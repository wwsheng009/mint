with open('runtime/paint/renderer.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
    for i in range(167, 182):
        print(f'Line {i+1}: {repr(lines[i])}')
