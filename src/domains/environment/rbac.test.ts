import { describe, expect, it } from "vitest";
import { canReadRole, canWriteRole } from "./rbac.js";

describe("RBAC helpers", () => {
  it("viewer can read not write", () => {
    expect(canReadRole("viewer")).toBe(true);
    expect(canWriteRole("viewer")).toBe(false);
  });
  it("editor can read and write", () => {
    expect(canReadRole("editor")).toBe(true);
    expect(canWriteRole("editor")).toBe(true);
  });
  it("undefined denies", () => {
    expect(canReadRole(undefined)).toBe(false);
    expect(canWriteRole(undefined)).toBe(false);
  });
});
