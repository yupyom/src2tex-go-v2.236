"""
{\ \centerline{\bf Docstring Test} }
This module demonstrates triple-quote comments.
"""

def hanoi(n, src, dst, tmp):
    '''Move n discs from src to dst using tmp.'''
    if n == 1:
        print(f"Move disc 1 from {src} to {dst}")
        return
    hanoi(n-1, src, tmp, dst)
    print(f"Move disc {n} from {src} to {dst}")
    hanoi(n-1, tmp, dst, src)

if __name__ == "__main__":
    hanoi(4, "A", "C", "B")   # {\ 4枚の disc を解く \hfill}
