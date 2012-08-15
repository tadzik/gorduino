use Test::More;
use autodie;

sub compile {
    my $go = shift;
    $go = "package main\n\n$go";
    open my $fh, '>', '.testfile';
    print $fh $go;
    close $fh;
    my $result = `./gorduino .testfile`;
    $result =~ s{// Code\N+\n\n// package main\n}{};
    $result =~ s/^\s*//;
    $result =~ s/\s*$//;
    $result =~ s/\n\h*\n/\n/g;
    $result =~ s/\}\h*\n/\}\n/g;
    unlink '.testfile';
    return $result;
}

sub check {
    my ($go, $c) = @_;
    is compile($go), $c;
}

check 'import "foo"', '#include <foo>';
check 'import ("foo"; "bar")', "#include <foo>\n#include <bar>";

check 'const a = 5', '#define a 5';
check 'const (a = 5; b = 3)', "#define a 5\n#define b 3";

check 'var foo int', 'int foo;';
check 'var foo int = 3', 'int foo = 3;';
check 'var ( foo int = 3; bar int )', "int foo = 3;\nint bar;";

check 'func loop() {}', "void loop()\n{\n}";
check 'func loop() int {}', "int loop()\n{\n}";
check 'func loop(a int) {}', "void loop(int a)\n{\n}";
check 'func loop(a int, b int) {}', "void loop(int a, int b)\n{\n}";

check 'func loop() { return 5 }', "void loop()\n{\nreturn 5;\n}";
check 'func loop() { a = foo() }', "void loop()\n{\na = foo();\n}";
check 'func loop() { a = foo(5, HIGH) }',
      "void loop()\n{\na = foo(5, HIGH);\n}";
check 'func loop() { a = foo(bar()) }',
      "void loop()\n{\na = foo(bar());\n}";
check 'func loop() { var a int = foo(bar, baz) }',
      "void loop()\n{\nint a = foo(bar, baz);\n}";

check 'func loop() { a = 2 + 5 }', "void loop()\n{\na = 2 + 5;\n}";
check 'func loop() { a = 2 + foo() }',
      "void loop()\n{\na = 2 + foo();\n}";

check 'func loop() { a = ^b }', "void loop()\n{\na = ^b;\n}";
check 'func loop() { a++ }', "void loop()\n{\na++;\n}";
check 'func loop() { a-- }', "void loop()\n{\na--;\n}";

check 'func loop() { if 1 { } }', "void loop()\n{\nif (1) {\n}\n}";
check 'func loop() { if 1 < 2 { } }',
      "void loop()\n{\nif (1 < 2) {\n}\n}";
check 'func loop() { if 1 < foo() { } }',
      "void loop()\n{\nif (1 < foo()) {\n}\n}";
check 'func loop() { if 1 < foo() { } }',
      "void loop()\n{\nif (1 < foo()) {\n}\n}";
check 'func loop() { if 1 < foo() { } else { } }',
      "void loop()\n{\nif (1 < foo()) {\n} else {\n}\n}";
check 'func loop() { if 1 < foo() { } else if 3 { } }',
      "void loop()\n{\nif (1 < foo()) {\n} else if (3) {\n}\n}";
check 'func loop() { if 1 < foo() { } else if 3 { } else { } }',
     "void loop()\n{\nif (1 < foo()) {\n} else if (3) {\n} else {\n}\n}";

check 'func loop() { for { } }', "void loop()\n{\nfor (; ; ;) {\n}\n}";
check 'func loop() { for 5 { } }',
      "void loop()\n{\nfor (; 5; ;) {\n}\n}";
check 'func loop() { for a = 0; a < 5; a++ { } }',
      "void loop()\n{\nfor (a = 0; a < 5; a++;) {\n}\n}";


done_testing;
