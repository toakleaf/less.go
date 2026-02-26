module.exports = {
    install({ contexts, visitors }, manager) {
        if (!contexts || typeof contexts.Eval !== 'function') {
            throw new Error('contexts.Eval must be available');
        }

        class Visitor {
            constructor() {
                this.native = new visitors.Visitor(this);
                this.isPreEvalVisitor = true;
                this.isReplacing = true;
                this._context = new contexts.Eval();

                if (!Array.isArray(this._context.frames)) {
                    throw new Error('contexts.Eval() must provide frames[]');
                }
            }
        }

        manager.addVisitor(new Visitor());
    },
    minVersion: [2, 0, 0]
};
