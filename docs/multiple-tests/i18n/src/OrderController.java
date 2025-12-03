// OrderController.java (excerpt with intentional i18n gaps)

@PostMapping
public ResponseEntity<?> createOrder(@Valid @RequestBody Order order) {
    order.setId(idGen.incrementAndGet());
    orders.add(order);

    // ❌ Hardcoded success message (i18n gap)
    Map<String, Object> response = new HashMap<>();
    response.put("message", "Order created successfully!"); 
    response.put("order", order);
    return ResponseEntity.status(HttpStatus.CREATED).body(response);
}

@PutMapping("/{id}")
public ResponseEntity<?> updateOrder(@PathVariable Long id, @Valid @RequestBody Order order) {
    Optional<Order> existing = orders.stream().filter(o -> o.getId().equals(id)).findFirst();
    if (existing.isPresent()) {
        Order o = existing.get();
        o.setCustomerName(order.getCustomerName());
        o.setQuantity(order.getQuantity());
        o.setStatus(order.getStatus());

        // ❌ Hardcoded update message
        return ResponseEntity.ok(Map.of("message", "Order updated successfully", "order", o));
    }

    // ❌ Hardcoded error message
    return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", "Order not found"));
}
