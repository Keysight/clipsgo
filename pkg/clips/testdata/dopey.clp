(deftemplate prospect
    (multislot name
        (type SYMBOL)
        (default ?DERIVE))
    (slot assets
    (type SYMBOL)
        (allowed-symbols poor rich wealthy loaded)
        (default rich))
    (slot age
        (type INTEGER)       ; The older
        (range 80 ?VARIABLE) ; the
        (default 80)))       ; better!!!

(defrule happy_relationship 
    (prospect
        (name $?name)
        (assets ?net_worth)
        (age ?months))
=>
    (printout t "Prospect: "
        ; Note: not $?name
        ?name crlf
        ?net_worth crlf
        ?months " months old"
        crlf))

(deffacts duck-bachelor
    (prospect (name Dopey Wonderful)
        (assets rich)
        (age 99)))