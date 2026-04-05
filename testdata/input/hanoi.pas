(* {\hrulefill }
 
{\ % beginning of TeX mode 

\centerline{\bf Towers of Hanoi (Pascal)}
 
\begin{quote}

This program gives an answer to the following famous problem (towers of
Hanoi).
There is a legend that when one of the temples in Hanoi was constructed,
three poles were erected and a tower consisting of 64 golden discs was
arranged on one pole, their sizes decreasing regularly from bottom to top.
The monks were to move the tower of discs to the opposite pole, moving
only one at a time, and never putting any size disc above a smaller one.
The job was to be done in the minimum numbers of moves. What strategy for
moving discs will accomplish this optimum transfer?
\end{quote}

% end of TeX mode }

{\hrulefill} *)

{ {\hrulefill\ hanoi.pas\ \hrulefill} }

program Hanoi;

const
    ARRAYSIZE = 8;                      { disc の数 }

var
    disc: array[0..2, 0..ARRAYSIZE-1] of integer;
                                        { disc に関するデータの置き場所 }
    ptr: array[0..2] of integer;        { disc 移動用ポインタ（インデックス）}
    counter: integer;                   { 移動回数カウンタ }

procedure InitArray;                    { disc に関するデータの初期化 }
var
    j: integer;
begin
    for j := 0 to ARRAYSIZE - 1 do
    begin
        disc[0][j] := ARRAYSIZE - j;
        disc[1][j] := 0;
        disc[2][j] := 0;
    end;
end;

procedure PrintResult;                  { 結果の表示 }
var
    i, j: integer;
begin
    counter := counter + 1;
    writeln('---#', counter, '---');
    for i := 0 to 2 do
    begin
        write('[', i, '] ');
        for j := 0 to ARRAYSIZE - 1 do
        begin
            if disc[i][j] <> 0 then
                write(disc[i][j], ' ')
            else
                break;
        end;
        writeln;
    end;
end;

procedure MoveOneDisc(i, j: integer);   { 1枚の disc を pole i から pole j に移動する }
begin
    ptr[i] := ptr[i] - 1;
    disc[j][ptr[j]] := disc[i][ptr[i]];
    ptr[j] := ptr[j] + 1;
    disc[i][ptr[i]] := 0;
end;

procedure MoveDiscs(n, i, j, k: integer);
                                        { \underline{\textsf{上から $n$ 枚目までの disc}}を、pole i から pole j に }
                                        { pole k を経由して、移動する }
begin
    if n >= 1 then
    begin
        MoveDiscs(n - 1, i, k, j);      { 関数 MoveDiscs() の中で、さらに自分自身 }
        MoveOneDisc(i, j);              { MoveDiscs() が使われている。このような }
        PrintResult;                    { 手法は、「再帰的呼びだし」といわれる。 }
        MoveDiscs(n - 1, k, j, i);
    end;
end;

{ {\par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center}

たとえば、関数 {\tt MoveDiscs(4, 0, 1, 2)} を呼び出すことは、
上図のような操作をすることに対応する。\hfill} }

begin
    counter := 0;
    ptr[0] := ARRAYSIZE;
    ptr[1] := 0;
    ptr[2] := 0;

    InitArray;
    MoveDiscs(ARRAYSIZE, 0, 1, 2);      { ARRAYSIZE 枚の disc を pole 0 から pole 1 に pole 2 }
                                        { を経由して、移動する }
end.
