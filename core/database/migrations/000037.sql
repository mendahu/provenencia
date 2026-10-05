-- S9-13: the `initial` name part type is retired. An initial is the part it
-- stands for (usually a given name) written short, so existing initial parts
-- become `given`. The name reconciler compares parts only with parts of the
-- same type, and "J." must meet "James". Audit history keeps the type that was
-- recorded at the time.

UPDATE name_value_parts SET type = 'given' WHERE type = 'initial';
