<!DOCTYPE html>
<!-- {\ \centerline{\bf Towers of Hanoi --- PHP Edition}} -->
<!--
{\ \begin{quote}
\noindent
This PHP script computes the solution to the Towers of Hanoi
puzzle server-side and renders the result as an HTML table.
The recursive algorithm requires exactly $2^n - 1$ moves to
transfer $n$ discs from one pole to another.  Run with
{\tt php hanoi.php > output.html} or serve via a web server.
\end{quote} }
-->
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>Hanoi Tower --- PHP Version</title>
    <style type="text/css">
        /* {\ \centerline{\bf Stylesheet}} */
        /* ページ全体のスタイル */
        body {
            font-family: "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        h1 { text-align: center; color: #333; }
        table {
            border-collapse: collapse;
            margin: 20px auto;
            background: #fff;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 6px 12px;
            text-align: center;
        }
        th {
            background-color: #4a90d9;
            color: #fff;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        .summary {
            text-align: center;
            margin: 10px 0;
            font-size: 14px;
            color: #666;
        }
    </style>
</head>
<body>
    <h1>Towers of Hanoi</h1>

<?php
/**
 * ハノイの塔 --- PHP版
 * サーバーサイドで再帰的に解を計算し、
 * HTML テーブルとして出力する
 */

$NUM_DISCS = 4;                         // {\ ディスクの枚数 \hfill}
$moves = array();                       // {\ 操作手順の配列 \hfill}

// {\ {\bf 再帰的解法} \hfill}
// {\ $n$ 枚のディスクを {\tt \$from} から}
// {\    {\tt \$to} へ {\tt \$via} を経由して移動 \hfill}
function hanoi($n, $from, $to, $via) {
    global $moves;
    if ($n <= 0) return;
    hanoi($n - 1, $from, $via, $to);    // {\ 上 $n-1$ 枚を退避 \hfill}
    $moves[] = array(                   // {\ 最大ディスクを移動 \hfill}
        'disc' => $n,
        'from' => $from,
        'to'   => $to
    );
    hanoi($n - 1, $via, $to, $from);    // {\ 退避分を目的地へ \hfill}
}

// {\ {\bf 移動回数の計算} \hfill}
// {\ [問題] この関数が $2^n - 1$ を返すことを}
// {\    確かめよ。 \hfill}
function countMoves($n) {
    if ($n == 0) return 0;
    return 1 + 2 * countMoves($n - 1);
}

/* 実行 */
hanoi($NUM_DISCS, 'A', 'C', 'B');
$expected = countMoves($NUM_DISCS);
?>

    <p class="summary">
        Discs: <?php echo $NUM_DISCS; ?> |
        Moves: <?php echo count($moves); ?> |
        Expected: <?php echo $expected; ?>
        (2<sup><?php echo $NUM_DISCS; ?></sup> - 1)
    </p>

    <!-- {\ {\bf 結果テーブル} \hfill} -->
    <table>
        <tr>
            <th>#</th>
            <th>Disc</th>
            <th>From</th>
            <th>To</th>
        </tr>
<?php foreach ($moves as $i => $m): ?>
        <tr>
            <td><?php echo $i + 1; ?></td>
            <td><?php echo $m['disc']; ?></td>
            <td><?php echo $m['from']; ?></td>
            <td><?php echo $m['to']; ?></td>
        </tr>
<?php endforeach; ?>
    </table>

    <script>
        // {\ \centerline{\bf JavaScript Section}}
        // {\ クライアント側の補助スクリプト \hfill}

        /* テーブル行のハイライト */
        var rows = document.querySelectorAll('tr');
        rows.forEach(function(row) {
            row.addEventListener('mouseover', function() {
                this.style.backgroundColor = '#e8f4fd';
            });
            row.addEventListener('mouseout', function() {
                this.style.backgroundColor = '';
            });
        });
    </script>
</body>
</html>
