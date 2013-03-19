<?php
	function getIntensity($img, $x, $y) {
		$rgb = imagecolorat($img, $x, $y);
		$r = ($rgb >> 16) & 0xFF;
		$g = ($rgb >> 8) & 0xFF;
		$b = $rgb & 0xFF;
		$a = floor(($r + $g + $b) / 3);
		return $a;
	}

	function calculateDev($img, $x, $y) {
		$a = getIntensity($img, $x+1, $y);
		$b = getIntensity($img, $x, $y);
		if($a == 0 || $b == 0) {
			return 0;
		}
		return abs($a-$b);
	}

	class MySimpleHeap extends SplHeap 
	{ 
	    public function  compare( $value1, $value2 ) { 
	        return ( $value1 - $value2 ); 
	    } 
	} 


	function imageDev($i, $minx, $maxx, $miny, $maxy) {
		$heap = new MySimpleHeap();
		if($i < 10) {
			$img = imagecreatefrompng('ffmpeg_temp/0'.$i.'.png');
		}
		else {
			$img = imagecreatefrompng('ffmpeg_temp/'.$i.'.png');
		}
		$mins = 0;
		$maxes = 0;
		for ($y = $miny; $y < $maxy; $y += 2) {
			for ($x = $minx; $x < $maxx; $x++) {
				$a = calculateDev($img, $x, $y);
				$heap->insert($a);
			}
		}
		for($i = 0; $i < 10; $i++) {
			$m = $heap->extract();
			$maxes += $m*$m*$m;
		}
		return $maxes;
	}

	function bestFit($minx, $maxx, $miny, $maxy) {
		$maxdev = 0;
		$maximage = 1;
		for($i = 1; $i < 25; $i ++) {
			$dev = imageDev($i, $minx, $maxx, $miny, $maxy);
			if($dev > $maxdev) {
				$maxdev = $dev;
				$maximage = $i - 1;
			}
		}
		return $maximage;
	}

	function imageFit($r, $c) {
		$img = imagecreatefrompng('ffmpeg_temp/01.png');
		$w = imagesx($img) / $c;
		$h = imagesy($img) / $r;
		echo '[';
		for($rr = 0; $rr < $r; $rr++) 
		{
			echo '[';
			for($cc = 0; $cc < $c; $cc++) {
				echo bestFit($cc * $w, $cc * $w + $w, $rr * $h, $rr * $h + $w);
				if($cc != $c - 1) {
					echo ',';
				}
			}
			echo ']';
			if ($rr != $r-1) {
				echo ',';
			}

			echo "<br>";
		}
		echo ']';
	}

	function microtime_float()
	{
	    list($usec, $sec) = explode(" ", microtime());
	    return ((float)$usec + (float)$sec);
	}

	set_time_limit(60);
	$time_start = microtime_float();
	imageFit(8,12);
	$time_end = microtime_float();
	echo "<br>".($time_end-$time_start);

?>