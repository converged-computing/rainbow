# A graph database backend provides the basic functions to interact with
# the graph database of rainbow, from a client perspective


class GraphBackend:
    """
    The GraphBackend is the base backend that primarily
    shows the abstract functions to be defined. Yes I'm not
    using abc. This is yet-another-hill... :)
    """

    def __init__(self, **options):
        """
        Create a new graph backend, accepting any options type.
        """
        # Set options as attributes
        for key, value in options.items():
            setattr(self, key, value)

    def satisfies(self, jobspec):
        """
        Determine if a jobspec can be satisfied
        """
        raise NotImplementedError
