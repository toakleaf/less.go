module.exports = {
    install({ tree, contexts, visitors }, manager) {
        if (typeof tree.Comment !== 'function') {
            throw new Error('tree.Comment must be available');
        }
        if (typeof tree.Unit !== 'function') {
            throw new Error('tree.Unit must be available');
        }

        // Smoke check constructor compatibility.
        new tree.Comment('/* plugin smoke */', false);

        class Visitor {
            constructor() {
                this.native = new visitors.Visitor(this);
                this.isPreEvalVisitor = true;
                this.isReplacing = true;
                this._context = new contexts.Eval();
                this._context.frames.push({
                    variable(name) {
                        if (name === '@replace') {
                            return {
                                value: new tree.Dimension(1, new tree.Unit(['px']))
                            };
                        }
                        return undefined;
                    }
                });
            }

            visitVariable(node) {
                const evaluated = node.eval(this._context);
                return evaluated || node;
            }
        }

        manager.addVisitor(new Visitor());
    },
    minVersion: [2, 0, 0]
};
